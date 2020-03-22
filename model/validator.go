package model

import (
	"encoding/json"
	"fmt"
)

// Validator ...
type Validator struct{}

// ValidateActionMsg ...
func (v *Validator) ValidateActionMsg(msg *string) error {
	actionMsg := &ActionMsg{}

	err := json.Unmarshal([]byte(*msg), actionMsg)

	if err != nil {
		return fmt.Errorf("Error when validating action message: %v", err)
	}

	if actionMsg.Action == "" {
		return fmt.Errorf("Field action of action message can not be empty")
	}

	return nil
}
