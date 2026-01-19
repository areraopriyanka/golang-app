package response

type ConfigResponse struct {
	ConfigName string      `json:"configName"`
	Type       string      `json:"type"`
	Value      interface{} `json:"value"`
}
