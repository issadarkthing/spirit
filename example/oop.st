
;; Although spirit is dynamic type language, defining members for class
;; requires type.
(defclass Human
  {:age types/Number
   :name types/String
   :sex types/String}
  
  (defmethod say-name [self]
    (print self.name)))


;; using the <- you can inherit from other class
(defclass Student <- Human
  {:id types/String}
  
  ;; defmethod needs to be defined in class
  ;; method requires one argument which is the instance itself
  (defmethod say-id [self]
    (print self.id)))
