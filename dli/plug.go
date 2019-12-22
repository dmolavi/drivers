package dli

type (
	Action struct {
		Type float64 `json:"type,omitempty"`
	}

	Child struct {
		ID         string `json:"id,omitempty"`
		State      int    `json:"state,omitempty"`
		Alias      string `json:"alias,omitempty"`
	}

	Sysinfo struct {
		Alias           string  `json:"alias,omitempty"`
		RelayState      int     `json:"relay_state,omitempty"`
		Children        []Child `json:"children,omitempty"`
	}

	System struct {
		Sysinfo Sysinfo `json:"get_sysinfo"`
	}
	Plug struct {
		System System `json:"system"`
	}
	Config struct {
		Address string `json:"address"`
	}
	CmdRelayState struct {
		System struct {
			RelayState struct {
				State int `json:"state"`
			} `json:"set_relay_state"`
		} `json:"system"`
		Context struct {
			Children []string `json:"child_ids,omitempty"`
		} `json:"context,omitempty"`
	}
)
