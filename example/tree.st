;;; Example of binary tree data structure

(defn from-list [coll]
  (reduce (fn [acc v] (insert-tree v acc)) nil coll)) 


(defn balanced [tree]
  (if (nil? tree)
    true
    (let [left (:left tree)
          right (:right tree)]
      (if (or 
            (not (balanced left)) 
            (not (balanced right))
            (not (zero? (- (height left) (height right)))))
        false
        true))))

(defn invert-tree [tree]
  (if (nil? tree)
      tree
      (let [x (:value tree)
            left (:left tree)
            right (:right tree)]
        (make-tree x (invert-tree right) (invert-tree left)))))
            
  

(defn insert-tree [val tree]
  (if (nil? tree)
    (make-tree val)
    (let [v (:value tree)
          left (:left tree)
          right (:right tree)]
      (if (= v val)
        (make-tree v left right)
        (if (< val v)
          (make-tree v (insert-tree val left) right)
          (make-tree v left (insert-tree val right)))))))

(defn in-order [tree]
  (if (nil? tree)
    '()
    (let [v (:value tree)
          left (:left tree)
          right (:right tree)]
      (concat (in-order left) (list v) (in-order right)))))

;; this is slower than in-order
(defn in-order-p [tree]
  (if (nil? tree)
    '()
    (let [v (:value tree)
          left (:left tree)
          right (:right tree)]
      (flat (wait-all (list (future (in-order left)) 
                            (future (list v)) 
                            (future (in-order right))))))))

(defn make-tree
  ([v]
   (make-tree v nil nil))
  ([v l r]
   {:value v :left l :right r}))

(defn max 
  ([a b]
   (if (> a b) a b))
  ([a b & z]
   (reduce max a (cons b z))))
   
(defn height [tree]
  (if (nil? tree)
    0
    (let [left (:left tree)
          right (:right tree)]
      (+ 1 (max (height left) (height right)))))) 

