package util

import (
	"strings"

	"github.com/google/uuid"
)

//生成UUID,无"-

func GenerateUUID() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")
}
