package server

import "time"

const SRC_REPOS_DIR_DEFAULT string = "/data/repos"
const SRC_REPOS_DESIRED_PERCENT_FREE_DEFAULT int = 10
const SRC_REPOS_JANITOR_INTERVAL_MIN_DEFAULT time.Duration = 1 * time.Minute
