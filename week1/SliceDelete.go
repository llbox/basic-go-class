package week1

// DeleteAt 切片删除指定下标元素
func DeleteAt[T any](sourceSlice []T, index int) []T {
	targetSlice := make([]T, 0, len(sourceSlice)-1)
	targetSlice = append(targetSlice, sourceSlice[:index]...)
	targetSlice = append(targetSlice, sourceSlice[index+1:]...)
	return targetSlice
}
