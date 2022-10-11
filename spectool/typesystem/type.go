package typesystem

import (
	"regexp"
)

type Type string

// DataType represents any Type that can be used for the properties of specifications such as enums, structs etc.
type DataType Type

const (
	Null       DataType = "null"
	Identifier DataType = "identifier"
	String     DataType = "string"
	Bool       DataType = "bool"
	Int        DataType = "int"
	Float      DataType = "float"
	Date       DataType = "date"
	DateTime   DataType = "dateTime"
	Char       DataType = "char"
	Array      DataType = "array"
	Map        DataType = "map"
)

func BuiltInDataTypes() []DataType {
	return []DataType{
		Null,
		Identifier,
		String,
		Bool,
		Int,
		Float,
		Date,
		DateTime,
		Char,

		// Containers
		Array,
		Map,
	}
}

const MapRegex = "^" + string(Map) + `\[(?P<KeyType>\w+)\](?P<ValueType>.*)`
const ArrayRegex = `^\[\](?P<ValueType>.*)`

// ExtractUserDefined searches the type for a user defined type and returns it.
// if the type does not contain any user defined type an empty string is returned.
// This method can be useful to find dependencies amongst field types.
func (dt DataType) ExtractUserDefined() DataType {
	if !dt.IsContainer() {
		if dt.IsUserDefined() {
			return dt
		}

		return ""
	}

	return dt.ContainerInfo().ValueType.ExtractUserDefined()
}

// IsUserDefined indicates if the type is a user-defined type or not.
func (dt DataType) IsUserDefined() bool {
	return !dt.IsBuiltIn()
}

// IsBuiltIn indicates if the type is a builtin type or not.
func (dt DataType) IsBuiltIn() bool {
	baseType := dt.BaseType()

	if dt.IsContainer() {
		return true
	}

	for _, t := range BuiltInDataTypes() {
		if t == baseType {
			return true
		}
	}

	return false
}

// IsArray indicates if a type is an array or not.
func (dt DataType) IsArray() bool {
	return dt.ArrayInfo() != ContainerTypeInfo{}
}

// IsMap indicates if a type is a map or not.
func (dt DataType) IsMap() bool {
	return dt.MapInfo() != ContainerTypeInfo{}
}

// IsContainer indicates if a type is a container type (e.g. Map or Array)
func (dt DataType) IsContainer() bool {
	return dt.IsMap() || dt.IsArray()
}

// Indicates if this data type is Null or not.
func (dt DataType) isNull() bool {
	return dt == Null
}

// BaseType returns the base-type of a container type, i.e. the right-most type
// If the type is not a container, will return the type itself.
func (dt DataType) BaseType() DataType {
	if dt.IsArray() {
		return dt.ContainerInfo().ValueType.BaseType()
	}

	if dt.IsMap() {
		return dt.ContainerInfo().ValueType.BaseType()
	}

	return dt
}

// ContainerInfo returns the info about a container. If the type is not a container type, this function will return an empty ContainerTypeInfo.
func (dt DataType) ContainerInfo() ContainerTypeInfo {

	mapInfo := dt.MapInfo()
	if mapInfo != (ContainerTypeInfo{}) {
		return mapInfo
	}

	arrayInfo := dt.ArrayInfo()
	if arrayInfo != (ContainerTypeInfo{}) {
		return arrayInfo
	}

	return ContainerTypeInfo{}
}

// MapInfo returns the info about a map. If the type is not a map, this function will return an empty ContainerTypeInfo.
func (dt DataType) MapInfo() ContainerTypeInfo {
	info := ContainerTypeInfo{}
	re := regexp.MustCompile(MapRegex)
	match := re.FindSubmatch([]byte(dt))

	if match == nil {
		return ContainerTypeInfo{}
	}

	for i, name := range re.SubexpNames() {
		if name == "KeyType" {
			info.KeyType = DataType(match[i])
		}
		if name == "ValueType" {
			info.ValueType = DataType(match[i])
		}
	}

	return info
}

// ArrayInfo returns the info about an array. If the type is not an array, this function will return an empty ContainerTypeInfo.
func (dt DataType) ArrayInfo() ContainerTypeInfo {
	re := regexp.MustCompile(ArrayRegex)
	match := re.FindSubmatch([]byte(dt))

	if match == nil {
		return ContainerTypeInfo{}
	}

	info := ContainerTypeInfo{}
	for i, name := range re.SubexpNames() {
		if name == "ValueType" {
			info.ValueType = DataType(match[i])
		}
	}
	info.KeyType = Int

	return info
}

type ContainerTypeInfo struct {
	KeyType   DataType
	ValueType DataType
}

// UndefinedValue This value is used internally to make the distinction between null/nil and undefined.
// For example a field's default value being set to undefined would not be equal to null.
// It would mean that there is not default value, as opposed to a value of null would mean exactly that.
const UndefinedValue = "__undefined__"
