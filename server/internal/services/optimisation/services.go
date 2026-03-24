package optimisation

import "fmt"

func InitService() (*Service, error) {
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create optimisation client: %s", err)
	}

	return client, nil
}
