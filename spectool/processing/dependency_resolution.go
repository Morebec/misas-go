// Copyright 2022 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processing

import (
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/pkg/errors"
	"strings"
)

type DependencySet map[spec.TypeName]struct{}

// diff Returns all elements that are in s and not in o. A / B
func (s DependencySet) diff(o DependencySet) DependencySet {

	diff := DependencySet{}

	for d := range s {
		if _, found := o[d]; !found {
			diff[d] = s[d]
		}
	}

	return diff
}

func (s DependencySet) TypeNames() []spec.TypeName {
	var typeNames []spec.TypeName

	for k := range s {
		typeNames = append(typeNames, k)
	}

	return typeNames
}

func NewDependencySet(dependencies ...spec.TypeName) DependencySet {
	deps := DependencySet{}
	for _, d := range dependencies {
		deps[d] = struct{}{}
	}

	return deps
}

// DependencyProvider are functions responsible for providing the dependencies of a list of Spec as a list of DependencyNode.
// Generally providers are specialized for a specific Type.
type DependencyProvider func(systemSpec spec.Spec, specs spec.Group) ([]DependencyNode, error)

type DependencyNode struct {
	Spec         spec.Spec
	Dependencies DependencySet
}

func (n DependencyNode) TypeName() spec.TypeName {
	return n.Spec.TypeName
}

func NewDependencyNode(spec spec.Spec, dependencies ...spec.TypeName) DependencyNode {
	return DependencyNode{Spec: spec, Dependencies: NewDependencySet(dependencies...)}
}

type DependencyGraph []DependencyNode

// Merge Allows merging this dependency graph with another one and returns the result.
func (g DependencyGraph) Merge(o DependencyGraph) DependencyGraph {
	var lookup = make(map[spec.TypeName]bool)
	var merge []DependencyNode

	for _, node := range g {
		merge = append(merge, node)
		lookup[node.TypeName()] = true
	}

	for _, node := range o {
		if _, found := lookup[node.TypeName()]; found {
			continue
		}
		merge = append(merge, node)
	}

	return NewDependencyGraph(merge...)
}

func NewDependencyGraph(nodes ...DependencyNode) DependencyGraph {
	return append(DependencyGraph{}, nodes...)
}

func (g DependencyGraph) Resolve() (ResolvedDependencies, error) {
	var resolved []DependencyNode

	// Look up of nodes to their typeName Names.
	nodesByTypeNames := map[spec.TypeName]DependencyNode{}

	// Map nodes to dependencies
	dependenciesByTypeNames := map[spec.TypeName]DependencySet{}
	for _, n := range g {
		nodesByTypeNames[n.TypeName()] = n
		dependenciesByTypeNames[n.TypeName()] = n.Dependencies
	}

	// The algorithm simply processes all nodes and tries to find the ones that have no dependencies.
	// When a node has dependencies, these dependencies are checked for being either circular or unresolvable.
	// If no unresolvable or circular dependency is found, the node is considered resolved.
	// And processing retries with the remaining dependent nodes.
	for len(dependenciesByTypeNames) != 0 {
		var typeNamesWithNoDependencies []spec.TypeName
		for typeName, dependencies := range dependenciesByTypeNames {
			if len(dependencies) == 0 {
				typeNamesWithNoDependencies = append(typeNamesWithNoDependencies, typeName)
			}
		}

		// If no nodes have no dependencies, in other words if all nodes have dependencies,
		// This means that we have a problem of circular dependencies.
		// We need at least one node in the graph to be independent for it to be potentially resolvable.
		if len(typeNamesWithNoDependencies) == 0 {
			// We either have circular dependencies or an unresolved dependency
			// Check if all dependencies exist.
			for typeName, dependencies := range dependenciesByTypeNames {
				for dependency := range dependencies {
					if _, found := nodesByTypeNames[dependency]; !found {
						return nil, errors.Errorf("spec with type name \"%s\" depends on an unresolved type name \"%s\"", typeName, dependency)
					}
				}
			}

			// They all exist, therefore, we have a circular dependencies.
			var circularDependencies []string
			for k := range dependenciesByTypeNames {
				circularDependencies = append(circularDependencies, string(k))
			}

			return nil, errors.Errorf("circular dependencies found between nodes \"%s\"", strings.Join(circularDependencies, "\", \""))
		}

		// All good, we can move the nodes that no longer have unresolved dependencies
		for _, nodeTypeName := range typeNamesWithNoDependencies {
			delete(dependenciesByTypeNames, nodeTypeName)
			resolved = append(resolved, nodesByTypeNames[nodeTypeName])
		}

		// Remove the resolved nodes from the remaining dependenciesByTypeNames.
		for typeName, dependencies := range dependenciesByTypeNames {
			diff := dependencies.diff(NewDependencySet(typeNamesWithNoDependencies...))
			dependenciesByTypeNames[typeName] = diff
		}
	}

	return append(ResolvedDependencies{}, resolved...), nil
}

// ResolvedDependencies represents an ordered list of DependencyNode that should be processed in that specific order to avoid
// unresolved types.
type ResolvedDependencies []DependencyNode

// DependenciesOfSpecTypeName returns the dependencies of a given spec.
func (d ResolvedDependencies) DependenciesOfSpecTypeName(tn spec.TypeName) []spec.TypeName {
	for _, s := range d {
		if s.Spec.TypeName == tn {
			return s.Dependencies.TypeNames()
		}
	}

	return nil
}
