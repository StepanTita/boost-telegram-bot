package config

type Generator interface {
	OpenAIApiToken() string
	Models() map[string]string
}

type generator struct {
	openAIApiToken string
	models         map[string]string
}

func NewGenerator(openAIAPIToken string, models map[string]string) Generator {
	return &generator{
		openAIApiToken: openAIAPIToken,
		models:         models,
	}
}

func (l generator) OpenAIApiToken() string {
	return l.openAIApiToken
}

func (l generator) Models() map[string]string {
	return l.models
}
