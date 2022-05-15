package test

import "fmt"

type Person struct {
	age  int
	name string
}

func newPerson(age int, name string) *Person {
	return &Person{
		age:  age,
		name: name,
	}
}

func (self *Person) SetAgeByPoint() {
	self.age = 99
}

func (self Person) SetAgeByValue() {
	self.age = 99
	fmt.Printf("THE age is %v\n", self.age)
}

func (self *Person) getAge() int {
	return self.age
}

func (self *Person) getName() string {
	return self.name
}

func main() {
	person := newPerson(18, "luo")
	person.SetAgeByValue()
	fmt.Printf("the age of person %v\n", person.getAge())
}
