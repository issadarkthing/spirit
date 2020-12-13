
### Immutability
As in Clojure, everything in Spirit is immutable which means you cannot re-assign value to variable. Everything that needs to be modified must be copied. However, this overhead can be reduced by using structural sharing. Thanks to [persistent](https://github.com/xiaq/persistent) library this data structure can be achieved similar to Clojure's PersistentVector and PersistentHashMap.



#### Class
```clojure
(defclass Person
    {:name ""
     :gender ""
     :age 0
     :nationality ""}

    (defmethod say-name [self]
        (print self.name))

    (defstatic new [name]
        (Person {:name name})))

(defclass Student <- Person
    {:id "0-28-1"}

    (defmethod greet-teacher [self]
        (print "Good morning teacher")))
```
Classes in Spirit are values which means that it can be passed around or bounded to **`Vars`**. The default value for members inside the class is just **`PersistentHashMap`** under the hood; same goes for methods and static members.


#### Objects
```clojure
(let [x (Person {:name "jiman"})]
    
    ;; prints jiman
    (x.say-name)

    (let [y (assoc x :name "bruhh")]

        ;; prints "bruhh" 
        (x.say-name)))
```
Objects also implemented as **`PersistentHashMap`**. You can use `assoc` to create a new object based on previous object. The overhead of it is low since it is using structural sharing that shares the same values with the previous object.
