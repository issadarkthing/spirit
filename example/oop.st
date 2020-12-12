
;; defining members inside class requires default value
(defclass Human
  {:age 0
   :name ""
   :sex ""}

  ;; defstatic is a static method of a class
  (defstatic init []
    (Human {:age 10, :name "jiman", :sex "male"}))
  
  (defmethod say-name [self]
    (print self.name)))


;; using the <- you can inherit from other class
(defclass Student <- Human
  {:id types/String}
  
  ;; defmethod needs to be defined in class
  ;; method requires one argument which is the instance itself
  (defmethod say-id [self]
    (print self.id)))
  
  
