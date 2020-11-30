(defclass Car
  {:name "toyota"
   :mileage 1000}
  
  (defmethod add-mileage [self mile]
    (assoc self :mileage (+ self.mileage mile)))
  
  (defstatic kind []
    "transportation"))
  

(defclass Human
  {:age 0
   :name ""
   :state "alive"}

  (defstatic code-name []
    "Homo Sapien")
  
  (defmethod aging [self]
    (assoc self :age (inc self.age)))
  
  (defmethod get-age [self]
    self.age)
  
  (defmethod add-age [self age]
    (+ (self.get-age) age)))


(defclass Student <- Human
  {:id ""
   :car (Car {})}
  
  (defmethod get-id [self]
    self.id)
  
  (defmethod get-factorial [self value]
    (if (= 1 value)
      value
      (* value (self.get-factorial (dec value))))))

(def student (Student {:id "10" :age 20}))

; accessing member
(assert (= student.id "10"))

; invoking method
(assert (= (student.get-id) "10"))

; inherit method and member
(assert (= (student.get-age) 20))
(assert (= student.age 20))

; use default value if member is not initialized
(assert (= "" student.name))

; access other method inside a method
(assert (= (student.add-age 10) 30))

; access nested member
(assert (= "toyota" student.car.name))

; access static method
(assert (= "transportation" (Car.kind)))

; inherit static method
(assert "inherit static method" 
        (= "Homo Sapien" (Student.code-name)))

(assert "recursive method" 
        (= 3628800 (student.get-factorial 10)))
