//go:build windows

package system

// platformCommands are the Windows equivalents of the systemd/journald
// diagnostics: sc for services, wevtutil for the event log, netstat for ports.
func platformCommands() []SafeCommand {
	return []SafeCommand{
		{
			Key:         "sc-query",
			Title:       "sc query",
			Description: "Показать запущенные службы Windows",
			Binary:      "sc",
			BuildArgs: func(string) ([]string, error) {
				return []string{"query"}, nil
			},
		},
		{
			Key:         "sc-query-service",
			Title:       "sc query <служба>",
			Description: "Показать состояние одной службы",
			NeedsArg:    true,
			ArgHint:     "Например: Spooler",
			Binary:      "sc",
			BuildArgs: func(arg string) ([]string, error) {
				name, err := requireServiceArg(arg)
				if err != nil {
					return nil, err
				}
				return []string{"query", name}, nil
			},
		},
		{
			Key:         "wevtutil-system",
			Title:       "wevtutil qe System /c:100",
			Description: "Показать последние события системного журнала",
			Binary:      "wevtutil",
			BuildArgs: func(string) ([]string, error) {
				return []string{"qe", "System", "/c:100", "/rd:true", "/f:text"}, nil
			},
		},
		{
			Key:         "wevtutil-application",
			Title:       "wevtutil qe Application /c:100",
			Description: "Показать последние события журнала приложений",
			Binary:      "wevtutil",
			BuildArgs: func(string) ([]string, error) {
				return []string{"qe", "Application", "/c:100", "/rd:true", "/f:text"}, nil
			},
		},
		{
			Key:         "netstat-ano",
			Title:       "netstat -ano",
			Description: "Показать слушающие порты и PID процессов",
			Binary:      "netstat",
			BuildArgs: func(string) ([]string, error) {
				return []string{"-ano"}, nil
			},
		},
	}
}
