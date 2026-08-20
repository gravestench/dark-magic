package host

import (
	"fmt"
	"strings"
)

const (
	dependencyUnvisited = iota
	dependencyVisiting
	dependencyVisited
)

// dependencyStartOrder performs a stable depth-first plan and rejects missing dependencies before startup begins.
func (host *Host) dependencyStartOrder() ([]string, error) {
	states := make(map[string]int, len(host.definitions))
	result := make([]string, 0, len(host.definitions))
	stack := make([]string, 0, len(host.definitions))

	var visit func(string) error

	visit = func(id string) error {
		switch states[id] {
		case dependencyVisited:
			return nil
		case dependencyVisiting:
			return fmt.Errorf("host: dependency cycle: %s -> %s", dependencyPath(stack), id)
		}

		definition, exists := host.definitions[id]
		if !exists {
			return fmt.Errorf("host: component %q is not registered", id)
		}

		// stack exists for actionable cycle errors; states alone can detect a cycle but cannot explain its path.
		states[id] = dependencyVisiting
		stack = append(stack, id)

		for _, dependency := range definition.DependsOn {
			if _, exists := host.definitions[dependency]; !exists {
				return fmt.Errorf(
					"host: component %q depends on unregistered component %q",
					id,
					dependency,
				)
			}

			if err := visit(dependency); err != nil {
				return err
			}
		}

		stack = stack[:len(stack)-1]
		states[id] = dependencyVisited
		// Post-order insertion puts every dependency before the component that consumes it.
		result = append(result, id)

		return nil
	}

	for _, id := range host.order {
		if err := visit(id); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// dependencyPath formats the active DFS stack without exposing graph implementation details in errors.
func dependencyPath(ids []string) string {
	return strings.Join(ids, " -> ")
}
