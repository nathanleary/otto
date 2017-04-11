package otto

import (
	"strconv"
	"sync"
)

func (self *_runtime) cmpl_evaluate_nodeProgram(node *_nodeProgram, eval bool) Value {
	if !eval {
		self.enterGlobalScope()
		defer func() {
			self.leaveScope()
		}()
	}

	self.cmpl_functionDeclaration(node.functionList)
	self.cmpl_variableDeclaration(node.varList)
	self.scope.frame.file = node.file
	return self.cmpl_evaluate_nodeStatementList(node.body)
}

func (self *_runtime) cmpl_call_nodeFunction(function *_object, stash *_fnStash, node *_nodeFunctionLiteral, this Value, argumentList []Value) Value {

	indexOfParameterName := make([]string, len(argumentList))
	// function(abc, def, ghi)
	// indexOfParameterName[0] = "abc"
	// indexOfParameterName[1] = "def"
	// indexOfParameterName[2] = "ghi"
	// ...

	argumentsFound := false

	l := len(node.parameterList)
	if l == 1 {

		index := 0
		if node.parameterList[index] == "arguments" {
			argumentsFound = true
		}
		value := Value{}
		if index < len(argumentList) {

			value = argumentList[index]
			//println(1)
			//go func() {
			indexOfParameterName[0] = node.parameterList[0]
			//}()
		}

		// strict = false

		self.scope.lexical.setValue(node.parameterList[index], value, false)

	} else if l > 1 {

		//for index, name := range node.parameterList {
		for index := 0; index < l; index++ {

			if node.parameterList[index] == "arguments" {
				argumentsFound = true
			}
			value := Value{}
			if index < len(argumentList) {
				value = argumentList[index]
				indexOfParameterName[index] = node.parameterList[index]
			}
			// strict = false
			self.scope.lexical.setValue(node.parameterList[index], value, false)
		}

	}

	if !argumentsFound {

		arguments := self.newArgumentsObject(indexOfParameterName, stash, len(argumentList))
		arguments.defineProperty("callee", toValue_object(function), 0101, false)
		stash.arguments = arguments
		// strict = false
		self.scope.lexical.setValue("arguments", toValue_object(arguments), false)
		//for index, _ := range argumentList {
		l := len(argumentList)
		if l == 1 {
			index := 0
			if index < len(node.parameterList) {
				//continue
			} else {
				indexAsString := strconv.FormatInt(int64(index), 10)
				arguments.defineProperty(indexAsString, argumentList[index], 0111, false)
			}
		} else if l > 1 {

			wg := sync.WaitGroup{}
			wg.Add(len(argumentList))
			for index := 0; index < len(argumentList); index++ {

				go func(index int) {
					defer wg.Done()
					if index < len(node.parameterList) {
						//continue
					} else {
						indexAsString := strconv.FormatInt(int64(index), 10)
						arguments.defineProperty(indexAsString, argumentList[index], 0111, false)
					}
				}(index)
			}
			wg.Wait()
		}
	}

	self.cmpl_functionDeclaration(node.functionList)
	self.cmpl_variableDeclaration(node.varList)

	result := self.cmpl_evaluate_nodeStatement(node.body)
	if result.kind == valueResult {
		return result
	}

	return Value{}
}

func (self *_runtime) cmpl_functionDeclaration(list []*_nodeFunctionLiteral) {

	executionContext := self.scope
	eval := executionContext.eval
	stash := executionContext.variable
	wg := sync.WaitGroup{}
	wg.Add(len(list))
	//for _, function := range list {
	for f := 0; f < len(list); f++ {

		name := list[f].name
		value := self.cmpl_evaluate_nodeExpression(list[f])
		if !stash.hasBinding(name) {
			go func(f int) {

				defer wg.Done()
				stash.createBinding(name, eval == true, value)
			}(f)
		} else {
			go func(f int) {

				defer wg.Done()
				// TODO 10.5.5.e
				stash.setBinding(name, value, false) // TODO strict
			}(f)
		}
	}
	wg.Wait()
}

func (self *_runtime) cmpl_variableDeclaration(list []string) {
	executionContext := self.scope
	eval := executionContext.eval
	stash := executionContext.variable
	//	wg := sync.WaitGroup{}
	//	wg.Add(len(list))
	//for _, name := range list {
	for f := 0; f < len(list); f++ {

		if !stash.hasBinding(list[f]) {

			//	go func(f int) {

			//	defer wg.Done()
			stash.createBinding(list[f], eval == true, Value{}) // TODO strict?
			//	}(f)
		} else {

			//	defer wg.Done()
		}

	}

	//wg.Wait()

}
