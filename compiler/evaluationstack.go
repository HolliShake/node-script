package main

type TEvaluationStack struct {
	stack []TValue
}

func CreateEvaluationStack() *TEvaluationStack {
	stack := new(TEvaluationStack)
	stack.stack = make([]TValue, 0)
	return stack
}

func (stack *TEvaluationStack) Push(value TValue) {
	stack.stack = append(stack.stack, value)
}

func (stack *TEvaluationStack) Pop() TValue {
	value := stack.stack[len(stack.stack)-1]
	stack.stack = stack.stack[:len(stack.stack)-1]
	return value
}

func (stack *TEvaluationStack) Peek() TValue {
	return stack.stack[len(stack.stack)-1]
}
