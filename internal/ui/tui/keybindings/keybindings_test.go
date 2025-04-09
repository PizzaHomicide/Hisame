package keybindings

import (
	"fmt"
	"testing"
)

func TestNoDuplicateKeyBindings(t *testing.T) {
	// Check each context individually
	for contextName, bindings := range ContextBindings {
		t.Run(fmt.Sprintf("Context_%s", contextName), func(t *testing.T) {
			keyToAction := make(map[string]Action)

			for _, binding := range bindings {
				// Check primary key
				if existingAction, exists := keyToAction[binding.KeyMap.Primary]; exists {
					t.Errorf("Duplicate key binding '%s' in context '%s': "+
						"first assigned to action '%s', then to '%s'",
						binding.KeyMap.Primary, contextName, existingAction, binding.Action)
				} else {
					keyToAction[binding.KeyMap.Primary] = binding.Action
				}

				// Check secondary key if it exists
				if binding.KeyMap.Secondary != "" {
					if existingAction, exists := keyToAction[binding.KeyMap.Secondary]; exists {
						t.Errorf("Duplicate key binding '%s' in context '%s': "+
							"first assigned to action '%s', then to '%s'",
							binding.KeyMap.Secondary, contextName, existingAction, binding.Action)
					} else {
						keyToAction[binding.KeyMap.Secondary] = binding.Action
					}
				}
			}
		})
	}
}
