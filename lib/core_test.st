; unicode characters and smileys
(def 🧠 "The Brain!")
(assert (= 🧠 "The Brain!"))

; basic functions
(assert (= 3 (eval '(number (+ 1 2)))))
(assert (true? true))
(assert (true? []))
(assert (not (true? false)))
(assert (not (true? nil)))

; sequence functions
(assert (seq? []))
(assert (not (seq? nil)))
(assert (= 1 (first [1 2 3 4])))
(assert (= [2 3 4] (next [1 2 3 4])))
(assert (= nil (next [])))
(assert (= [1 2 3 4] (cons 1 [2 3 4])))
(assert (= [1 2 3 4] (conj [1 2 3] 4)))


; threading macros
(assert (= (-> 1 (cons [2 3 4])) [1 2 3 4]))
(assert (= (-> 5 (+ 3) (/ 2) (- 1)) 3))
(assert (= (->>  1 (conj [2 3 4])) [2 3 4 1]))
(assert (= (->> (range-vec 1 11)
                (map #(* %1 %1)) 
                (filter even?) 
                (take 2) 
                (reduce +))
           20))
(assert (= (->> 12
                (list)
                (cons 20)
                (map inc))
           '(21 13)))

; basic math operators
(assert (= 3 (+ 1 2)))
(assert (= 3 (- 5 2)))
(assert (= -5 (- 5)))
(assert (= 10 (* 5 2)))
(assert (= 5 (/ 10 2)))
(assert (= 0.5 (/ 1 2)))
(assert (> 10 9 8 7 6 1 -1 -10))
(assert (< -10 1 2 3 4 10 23.32423432 100000))
(assert (>= 10 10 10 9 8 7 7 7 5))
(assert (<= -1.5 -1 0 0  0  0 0 0 1 2 3 4 5))

; type initialization functions
(assert (= #{} (set [])))
(assert (= #{1 2 3} (set [1 1 2 2 3])))
(assert (= [] (vector)))
(assert (= [1 2 ["hello"]] (vector 1 2 ["hello"])))
(assert (= () (list)))
(assert (= '(1 [] ["hello"] "hello") (list 1 [] ["hello"] "hello")))
(assert (= "" (str nil)))
(assert (= "" (str)))
(assert (= "1" (str 1)))
(assert (= "hello-bob" (str "hello-" "bob")))

; ; type checking functions
(assert (number? 10))
; (assert (not (int? 10.0)))
(assert (string? ""))
(assert (not (string? nil)))
(assert (boolean? true))
(assert (boolean? false))
(assert (not (boolean? nil)))
(assert (vector? []))
(assert (not (vector? nil)))
(assert (= (type [])(type [1 2 3])))
(assert (symbol? 'hello))
(assert (= '(1 2 3 4 5 6) (concat [1 2 3] '(4 5 6))))
(assert (= 10 (-> 5 (+ 5) number)))
(assert (= "hello" (-> 0 ["hello" 1 2 #{}])))


;; hashmap tests
(assert (= {:a 1} {:a 1}))
(assert (= (assoc {:a 1} :b 2) {:a 1 :b 2}))
(assert (= "jiman" (:name {:name "jiman"})))
(assert (= "jiman" ({:name "jiman"} :name)))

; ; simple function definition
(def dec (fn* [i] (number (- i 1))))
(assert (= 9 (dec 10)))

; ; simple recursive function with variadic args
(def down-range (fn* down-range [start & args]
                     (if (> start 0)
                       (cons start (down-range (dec start)))
                       [0])))
(assert (= '(5 4 3 2 1 0) (down-range 5)))

; ; complex recursive function
(def reverse-x (fn* reverse-x [coll]
                    (if (not (seq? coll))
                      (throw "argument must be a sequence"))
                    (if (nil? (next coll))
                      [(first coll)]
                      (let [first-value   (first coll)
                            reversed      (reverse-x (next coll))]
                        (conj reversed first-value)))))
(assert (= '(5 4 3 2 1) (reverse-x '(1 2 3 4 5))))

(def fib (fn* fib [n]
              (if (> n 1)  ; if n=0 or n=1 return n
                (+ (fib (- n 1)) (fib (- n 2)))
                n)))
(assert (= 2584 (fib 18)))

; ; multi arity function
(def greet (fn* greet
                ([] "Hello!")
                ([name] (str "Hello " name "!"))
                ([prefix name] (str prefix " " name "!"))))
(assert (= "Hello!" (greet)))
(assert (= "Hello Bob!" (greet "Bob")))
(assert (= "Hi Bob!" (greet 'Hi 'Bob)))

; ; tests for special forms
(def nested-special-forms (fn* defn [name args & body]
                               `(def ~name (fn* ~args (do (quote ~body))))))

(assert (= '(def hello (fn* [arg] (do (quote (arg)))))
           (nested-special-forms 'hello '[arg] 'arg)))

(assert (= "Hello Bob!"
           (let [name "Bob"]
             (str "Hello " name "!"))))

(def sum-through-let (<> + [1 2 3 4 5]))
(assert (= (+ 1 2 3 4 5) sum-through-let))


(defn square [num] (* num num))
(defn positive? [num] (> num 0))

(assert (= '(0 1 2) (range 3)))
(assert (= '(4 5 6 7 8) (range 4 9)))
(assert (= '(4 6 8) (range 4 9 2)))

(assert (= [1 4 9 25] (map square [1 2 3 5])))
(assert (= '(1 4 9 25) (map square '(1 2 3 5))))
(assert (= '(10 1 12 3 14 5) (map-indexed 
                               (fn [x i] 
                                 (if (even? i)
                                   (+ 10 x)
                                   x)) 
                               (range 6))))


(assert (= '(2 5 4 1) (filter positive? [-1 0 -2 2 5 4 1])))
(assert (= '(0 2 4) (filter-indexed (fn [x i] (even? i)) (range 6))))

(assert (= 45 (reduce + (range 10))))
(assert (= 55 (reduce + 10 (range 10))))
(assert (= 6 (reduce + [1 2 3])))
(assert (= 26 (reduce + 20 [1 2 3])))
(assert (= '(0 2 4 6 8) (reduce-indexed 
                          (fn [acc v i]
                            (if (even? i)
                              (conj acc v)
                              acc)) '() (range 10))))


(assert (= 10 (hoist-test)))
(defn hoist-test [] 10)


(assert (= (rest [1]) []))
(assert (= (rest '(1)) '()))

(assert (= (drop 2 '(1 2 3)) '(3)))
(assert (= (drop 2 [1 2 3]) [3]))


(assert (some even? '(1 3 5 8)))
(assert (every even? '(2 4 6 8 10)))

(assert (= 8 (find even? '(1 3 5 8 7))))
(assert (= nil (find even? '(1 3 5 11 7))))


(assert (= [1 2 3] (flat [1 [2 [3]]])))
(assert (= '(1 2 3) (flat '(1 (2 (3))))))


(assert (= 10 (abs -10)))
(assert (= 10 (abs 10)))

(assert (= [1 2 3] (range-vec 1 4)))
(assert (= 10 (tap (+ 5 5))))

(assert (= (last '(1 2 3)) 3))
(assert (= (last [1 2 3]) 3))

(assert (= (reverse [1 2 3]) [3 2 1]))
(assert (= (reverse '(1 2 3)) '(3 2 1)))

(defmacro splice-test [& exprs]
  `(list ~@exprs))

(assert (= (splice-test 1 2 3) '(1 2 3)))

(assert (= #[1 100] (lazy-range 1 100)))
(assert (= '(0 1 2) (take 3 #[1000000000000])))
(assert (= '(5 7 9 11 13) (take 5 #[5 100000 2])))
(assert (= nil (first #[0])))


;; syntax lambda

; omit first paramater
(assert (= 11 (reduce #(+ 1 %2) '(1 2 10))))

(defn pass-four [f]
  (f 1 2 3 4))

; omit first three paramater
(assert (= [4] (pass-four #(vector %4))))

; nested form
(assert (= 6 (pass-four #(+ %4 (+ %1 %1)))))


