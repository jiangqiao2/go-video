package assert

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

var (
	initStack   []string
	stackMutex  sync.Mutex
)

// NotCircular 检查循环依赖
func NotCircular() {
	stackMutex.Lock()
	defer stackMutex.Unlock()

	// 获取调用者信息
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return
	}

	caller := fmt.Sprintf("%s:%d", getShortFileName(file), line)

	// 检查是否已经在初始化栈中
	for _, existing := range initStack {
		if existing == caller {
			panic(fmt.Sprintf("Circular dependency detected: %s\nInit stack: %s", 
				caller, strings.Join(initStack, " -> ")))
		}
	}

	// 添加到初始化栈
	initStack = append(initStack, caller)

	// 使用defer在函数返回时移除
	defer func() {
		stackMutex.Lock()
		defer stackMutex.Unlock()
		if len(initStack) > 0 {
			initStack = initStack[:len(initStack)-1]
		}
	}()
}

// NotNil 检查对象不为空
func NotNil(obj interface{}) {
	if obj == nil {
		_, file, line, _ := runtime.Caller(1)
		panic(fmt.Sprintf("Object is nil at %s:%d", getShortFileName(file), line))
	}
}

// NotEmpty 检查字符串不为空
func NotEmpty(str string) {
	if str == "" {
		_, file, line, _ := runtime.Caller(1)
		panic(fmt.Sprintf("String is empty at %s:%d", getShortFileName(file), line))
	}
}

// True 检查条件为真
func True(condition bool, message string) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		panic(fmt.Sprintf("Assertion failed: %s at %s:%d", message, getShortFileName(file), line))
	}
}

// False 检查条件为假
func False(condition bool, message string) {
	if condition {
		_, file, line, _ := runtime.Caller(1)
		panic(fmt.Sprintf("Assertion failed: %s at %s:%d", message, getShortFileName(file), line))
	}
}

// Equal 检查两个值相等
func Equal(expected, actual interface{}, message string) {
	if expected != actual {
		_, file, line, _ := runtime.Caller(1)
		panic(fmt.Sprintf("Assertion failed: %s. Expected: %v, Actual: %v at %s:%d", 
			message, expected, actual, getShortFileName(file), line))
	}
}

// NotEqual 检查两个值不相等
func NotEqual(expected, actual interface{}, message string) {
	if expected == actual {
		_, file, line, _ := runtime.Caller(1)
		panic(fmt.Sprintf("Assertion failed: %s. Values should not be equal: %v at %s:%d", 
			message, expected, getShortFileName(file), line))
	}
}

// getShortFileName 获取短文件名
func getShortFileName(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return fullPath
}

// ClearStack 清空初始化栈（用于测试）
func ClearStack() {
	stackMutex.Lock()
	defer stackMutex.Unlock()
	initStack = nil
}