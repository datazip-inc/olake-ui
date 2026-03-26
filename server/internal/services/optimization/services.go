package optimization

import "fmt"

func InitService() (*Service, error) {
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create optimization client: %s", err)
	}

	return client, nil
}
