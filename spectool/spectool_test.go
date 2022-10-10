package spectool

import (
	"context"
	"testing"
)

func TestSpectool(t *testing.T) {
	err := Spectool(
		context.Background(),
		"./spec_testdata/system.spec.yaml",
		LoadSpecs(
			SystemSpecDeserializer(),
			CommandDeserializer(),
			QueryDeserializer(),
			EventDeserializer(),
			StructDeserializer(),
			EnumDeserializer(),
			HTTPEndpointDeserializer(),
		),
		LintSpecs(
			// Common
			SpecificationsMustNotHaveUndefinedTypes(),
			SpecificationMustNotHaveUndefinedTypeNames(),
			SpecificationsMustNotHaveDuplicateTypeNames(),
			SpecificationsMustHaveDescriptions(),
			SpecificationsMustHaveLowerCaseTypeNames(),
			SpecificationsShouldFollowNamingConvention(),
			// Structs
			// Enums
			EnumBaseTypeShouldBeSupportedEnumBaseType(),

			// Commands
			// Queries
			// Events
			EventsMustHaveDateTimeField(),

			// HTTP Endpoints
			HTTPEndpointsShouldFollowNamingConvention(),
			HTTPEndpointPathsShouldStartWithForwardSlash(),
			HTTPEndpointPathsShouldNotEndWithForwardSlash(),
			HTTPEndpointPathsShouldBeUnique(),
			HTTPEndpointPathShouldBeLowercase(),
			HTTPEndpointsShouldHaveEitherGETorPOSTMethod(),
			HTTPEndpointsWithCommandRequestTypeMustHaveMethodPOST(),
			HTTPEndpointsWithQueryRequestTypeMustHaveMethodGET(),
			HTTPEndpointResponseShouldHaveValidStatusCode(),
		),
		ResolveDependencies(
			SystemDependencyProvider(),
			CommandDependencyProvider(),
			QueryDependencyProvider(),
			EventDependencyProvider(),
			StructDependencyProvider(),
			EnumDependencyProvider(),
			HTTPEndpointDependencyProvider(),
		),
		GoProcessor(),
		WriteOutputFiles(),
	)
	if err != nil {
		panic(err)
	}
}
