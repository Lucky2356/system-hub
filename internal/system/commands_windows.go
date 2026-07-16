//go:build windows

package system

import "github.com/Lucky2356/system-hub/internal/i18n"

// platformCommands are the Windows equivalents of the systemd/journald
// diagnostics: sc for services, wevtutil for the event log, netstat for ports.
// The list is rebuilt on every call, so a language switch reaches it.
func platformCommands() []SafeCommand {
	return []SafeCommand{
		{
			Key:         "sc-query",
			Title:       "sc query",
			Description: i18n.T("Show running Windows services"),
			Binary:      "sc",
			BuildArgs: func(string) ([]string, error) {
				return []string{"query"}, nil
			},
		},
		{
			Key:         "sc-query-service",
			Title:       "sc query " + i18n.T("<service>"),
			Description: i18n.T("Show the state of a single service"),
			NeedsArg:    true,
			ArgHint:     i18n.Tf("For example: %s", "Spooler"),
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
			Description: i18n.T("Show the latest system event log entries"),
			Binary:      "wevtutil",
			BuildArgs: func(string) ([]string, error) {
				return []string{"qe", "System", "/c:100", "/rd:true", "/f:text"}, nil
			},
		},
		{
			Key:         "wevtutil-application",
			Title:       "wevtutil qe Application /c:100",
			Description: i18n.T("Show the latest application event log entries"),
			Binary:      "wevtutil",
			BuildArgs: func(string) ([]string, error) {
				return []string{"qe", "Application", "/c:100", "/rd:true", "/f:text"}, nil
			},
		},
		{
			Key:         "netstat-ano",
			Title:       "netstat -ano",
			Description: i18n.T("Show listening ports and process IDs"),
			Binary:      "netstat",
			BuildArgs: func(string) ([]string, error) {
				return []string{"-ano"}, nil
			},
		},
	}
}
