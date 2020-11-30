
### Object Oriented System
In Spirit, you can define `Class` and `Object` like other OOP languages with
only difference that both `Class` and `Object` are immutable.

#### Defining class
Defining a class in Spirit requires you to initiliaze default value for each
members. This helps to reduce bugs in most dynamic type system.
```clojure
(defclass Human
    {:age 0})
```

#### Inheritance
You can also inherit from other `Class` using `<-`
```clojure
(defclass Student <- Human
    {:id ""})
```

#### Defining method and static methods
`defmethod` need to be defined in `defclass`. Each method receives the instance 
of the class as the first argument. First argument can be named other than `self`.
`defstatic` is defined like defmethod but does not receive `self` argument.
```clojure
(defclass Class-Rep <- Student
    {:class ""}
    
    (defmethod report [self]
        (print (str self.class " is doing fine.")))

	(defstatic id []
		"012-321"))
```

#### Class instantiation
To instantiate a class. You can use the `class` itself and pass hashmap
containing the members and values for the `class`. Note that if you pass member
that is not defined it will throw error and unlikewise not passing value for
member will be initiliazed to `nil`. 
```clojure
(def student (Class-Rep {:id "007" :age 20}))
```

#### Member access and method invocation
Both member access and method invocation can use `.` operator just like in `java`
```clojure
;; calling a method
(student.report)

;; access member
student.name

;; calling static method
(Student.id)
```
