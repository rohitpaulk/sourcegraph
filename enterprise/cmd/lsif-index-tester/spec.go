package main

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position
	End   Position
}

type DefinitionRequest struct {
	TextDocument string   `json:"textDocument"`
	Position     Position `json:"position"`
}

type DefinitionResponse struct {
	TextDocument string `json:"textDocument"`
	Range        Range
}

type DefinitionTest struct {
	Request  DefinitionRequest  `json:"request"`
	Response DefinitionResponse `json:"response"`
}

type LsifTest struct {
	Definitions map[string]DefinitionTest `json:"textDocument/definition"`
}
