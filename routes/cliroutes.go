package routes

import (
	"database/sql"
	"fmt"
	"jobqueue/modules"
	"strconv"

	"jobqueue/controllers"
)

func isValidState(s string) bool {
	return s == modules.STATE_WAITING ||
		s == modules.STATE_IN_PROGRESS ||
		s == modules.STATE_COMPLETED ||
		s == modules.STATE_FAILED
}

func Dispatch(db *sql.DB, args []string) {
	if len(args) == 0 {
		Help()
		return
	}

	switch args[0] {

	case "createjob":
		controllers.CreateJobController(db, args[1:])

	case "startworker":
		cnt := 1
		if len(args) >= 2 {
			v, err := strconv.Atoi(args[1]) // <-- THIS checks number-only
			if err != nil || v <= 0 {
				fmt.Println("wrong :worker count ")
				Help()
				return
			}
			cnt = v
		}
		controllers.StopWorkerController(db, cnt)

	case "stopworker":
		controllers.StopWorkerController(db)

	case "status":
		controllers.GetStatusController(db)

	case "listjob":
		if len(args) < 2 {
			fmt.Println("wrong: state not given")
			Help()
			return
		}

		state := args[1]

		if !isValidState(state) {
			fmt.Println("wrong: invalid state ->", state)
			fmt.Println("allowed states: WAITING | IN_PROGRESS | COMPLETED | FAILED")
			return
		}

		controllers.ListJobsController(db, args[1:])

	default:
		fmt.Println("wrong command")
		Help()
	}
}

func Help() {
	fmt.Println(`Commands:
  createjob "<command>"
  startworker 3
  stopworker
  status
  listjob <WAITING|IN_PROGRESS|COMPLETED|FAILED>`)
}
