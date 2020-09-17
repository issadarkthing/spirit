(ns 'simple)


(defn partial-range [min]
  (range min *max*))

(defn simple-greet [name]
  (print (str "hi " name)))
