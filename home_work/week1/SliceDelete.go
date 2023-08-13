package week1

import "errors"

// DeleteAt 切片删除指定下标元素
func DeleteAt[T any](sourceSlice []T, index int) ([]T, error) {
	targetSlice := make([]T, 0, len(sourceSlice)-1)
	if index < 0 || index > len(sourceSlice)-1 {
		return targetSlice, errors.New("invalid index error")
	}

	targetSlice = append(targetSlice, sourceSlice[:index]...)
	targetSlice = append(targetSlice, sourceSlice[index+1:]...)
	return targetSlice, nil
}
