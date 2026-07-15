package models

type Windows struct {
	Items []Window `json:"data"`
}

type Window struct {
	Id             int64  `json:"id"`
	IsLastFocused  bool   `json:"isLastFocused"`
	Focused        bool   `json:"focused"`
	Type           string `json:"type"`
	State          string `json:"state"`
	Incognito      bool   `json:"incognito"`
	TabCount       int    `json:"tabCount"`
	ActiveTabTitle string `json:"activeTabTitle"`
	ActiveTabUrl   string `json:"activeTabUrl"`
	Top            int64  `json:"top"`
	Left           int64  `json:"left"`
	Width          int64  `json:"width"`
	Height         int64  `json:"height"`
}
