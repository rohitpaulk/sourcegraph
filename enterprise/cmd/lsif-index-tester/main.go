package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"
	"github.com/pelletier/go-toml"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsif/correlation"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// TODO:
//   We need to check for lsif-clang, lsif-validate at start time.

type ProjectResult struct {
	success bool
	usage   UsageStats
	output  string
}

type IndexerResult struct {
	usage  UsageStats
	output []byte
}

type UsageStats struct {
	// Memory usage in kilobytes by child process.
	memory int64
}

func main() {
	logging.Init()
	trace.Init(false)

	log15.Root().SetHandler(log15.StdoutHandler)

	var directory string
	flag.StringVar(&directory, "dir", ".", "The directory to run the test harness over")

	var indexer string
	flag.StringVar(&indexer, "indexer", "", "The name of the indexer that you want to test")

	var monitor bool
	flag.BoolVar(&monitor, "monitor", true, "Whether to monitor and log stats")

	flag.Parse()

	if indexer == "" {
		log.Fatalf("Indexer is required. Pass with --indexer")
	}

	log15.Info("Starting Execution: ", "directory", directory, "indexer", indexer)

	// w, _ := os.Create("monitor.csv")
	// testContext = context.WithValue(testContext, "output", w)

	testContext := context.Background()
	testContext = context.WithValue(testContext, "monitor", monitor)

	err := testDirectory(testContext, indexer, directory)
	if err != nil {
		log.Fatalf("Failed with: %s", err)
	}
}

func testDirectory(ctx context.Context, indexer string, directory string) error {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, f := range files {
		log15.Info("Running test for: ", "file", f.Name())
		projResult, err := testProject(ctx, indexer, directory+"/"+f.Name())
		if err != nil {
			fmt.Printf("Project %+v", projResult)
			log.Fatalf("Project '%s' failed to complete", f.Name())
		}

		log15.Info("Project result:", "success", projResult.success)
		if !projResult.success {
			log.Fatalf("Project '%s' failed test", f.Name())
		}
	}

	return nil
}

func testProject(ctx context.Context, indexer, project string) (ProjectResult, error) {

	var result IndexerResult
	var err error

	if indexer == "lsif-clang" {
		log15.Debug("... Starting lsif clang")
		result, err = testLsifClang(ctx, project)
		if err != nil {
			return ProjectResult{
				success: false,
				usage: UsageStats{
					memory: -1,
				},
				output: string(result.output),
			}, err
		}
	}

	log15.Debug("... \t Resource Usage:", "usage", result.usage)

	output, err := validateDump(project)
	if err != nil {
		fmt.Println("Not valid")
		return ProjectResult{
			success: false,
			usage:   result.usage,
			output:  string(output),
		}, err
	} else {
		log15.Debug("... Validated dump.lsif")
	}

	bundle, err := readBundle(1, project)
	// fmt.Printf("Bundle: %+v\n", bundle)
	// if err != nil {
	// 	return []byte{}, err
	// }
	// // fmt.Printf("Bundle: %+v\n", bundle)

	validateTestCases(project, bundle)

	return ProjectResult{
		success: true,
		usage:   result.usage,
		output:  string(output),
	}, nil
}

// TODO this is pretty dumb... we should find some way that isn't hard coded in here to run it.
// But for now I'm going to do it this way.
func testLsifClang(ctx context.Context, project string) (IndexerResult, error) {
	output, err := generateCompileCommands(project)
	if err != nil {
		return IndexerResult{
			usage:  UsageStats{memory: -1},
			output: output,
		}, err
	} else {
		log15.Debug("... Generated compile_commands.json")
	}

	output, usage, err := runLsifClang(ctx, project)
	if err != nil {
		log.Println(output)
		log.Fatal(err)
	} else {
		log15.Debug("... Generated dump.lsif")
	}

	return IndexerResult{
		usage:  usage,
		output: output,
	}, err
}

func generateCompileCommands(directory string) ([]byte, error) {
	cmd := exec.Command("./get_compile_commands.sh")
	cmd.Dir = directory

	return cmd.CombinedOutput()
}

func runLsifClang(ctx context.Context, directory string) ([]byte, UsageStats, error) {
	// TODO: We should add how long it takes to generate this.

	cmd := exec.Command("lsif-clang", "compile_commands.json")
	cmd.Dir = directory

	log15.Debug("... Generating dump.lsif")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, UsageStats{memory: -1}, err
	}

	sysUsage := cmd.ProcessState.SysUsage()
	mem, _ := MaxMemoryInKB(sysUsage)
	// fmt.Println("Memory Usage:", mem, "kB")
	// fmt.Println("User CPU", sysUsage.Utime)

	return output, UsageStats{memory: mem}, err
}

func validateDump(directory string) ([]byte, error) {
	// TODO: Eventually this should use the package, rather than the installed module
	//       but for now this will have to do.
	cmd := exec.Command("lsif-validate", "dump.lsif")
	cmd.Dir = directory

	return cmd.CombinedOutput()
}

func validateTestCases(directory string, bundle *correlation.GroupedBundleDataMaps) {
	doc, err := ioutil.ReadFile(directory + "/test.toml")
	if err != nil {
		log15.Warn("No file exists here")
		return
	}

	testCase := LsifTest{}
	toml.Unmarshal(doc, &testCase)

	for _, definitionRequest := range testCase.Definitions {
		path := definitionRequest.Request.TextDocument
		line := definitionRequest.Request.Position.Line
		character := definitionRequest.Request.Position.Character

		results, err := correlation.Query(bundle, path, line, character)

		if err != nil {
			log.Fatalf("Failed query: %s", err)
		}

		if len(results) != 1 {
			log.Fatalf("Had too many results: %v", results)
		}

		definitions := results[0].Definitions

		if len(definitions) > 1 {
			log.Fatalf("Had too many definitions: %v", definitions)
		} else if len(definitions) == 0 {
			log.Fatalf("Found no definitions: %v", definitions)
		}

		response := transformLocationToResponse(definitions[0])
		if diff := cmp.Diff(response, definitionRequest.Response); diff != "" {
			log.Fatalf("Bad diffs: %s", diff)
		}
	}

	log15.Info("Passed tests")
}

func transformLocationToResponse(location lsifstore.LocationData) DefinitionResponse {
	return DefinitionResponse{
		TextDocument: location.URI,
		Range: Range{
			Start: Position{
				Line:      location.StartLine,
				Character: location.StartCharacter,
			},
			End: Position{
				Line:      location.EndLine,
				Character: location.EndCharacter,
			},
		},
	}

}

func getWriter(ctx context.Context) *os.File {
	return ctx.Value("output").(*os.File)
}
