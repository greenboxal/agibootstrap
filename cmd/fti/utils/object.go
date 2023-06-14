package utils

import (
	"fmt"
	"path/filepath"
)

// Object is the struct representation of an object in FTI
type Object struct {
	Hash     string
	ChunkMap map[string]Chunk
}

// Chunk is the struct representation of a chunk in FTI
type Chunk struct {
	Size      int
	Overlap   int
	Embedding []float64
}

// CreateObject creates a new object and saves it to disk
func CreateObject(path string, obj Object) error {
	// TOxDO: Serialize the object into a format suitable for storage (e.g., JSON, binary, etc.)
	serializedObj := obj

	// Write serialized object to file
	err := WriteFile(filepath.Join(path, obj.Hash+".bin"), serializedObj)

	if err != nil {
		return fmt.Errorf("failed to create object: %v", err)
	}

	return nil
}

// RetrieveObject retrieves an object from disk
func RetrieveObject(path string, hash string) (Object, error) {
	var obj Object

	// Read object file
	content, err := ReadFile(filepath.Join(path, hash+".bin"))

	if err != nil {
		return obj, fmt.Errorf("failed to read object file: %v", err)
	}

	// TOxDO: Deserialize the content into an Object
	obj = content // Deserialization result

	return obj, nil
}
