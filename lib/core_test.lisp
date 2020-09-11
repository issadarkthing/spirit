; unicode characters and smileys
(def 🧠 "The Brain!")
(assert (= 🧠 "The Brain!"))

; ; basic functions
(assert (= 3 (eval '(int (+ 1 2)))))
(assert (true? true))
(assert (true? []))
(assert (not (true? false)))
(assert (not (true? nil)))

; ; sequence functions
(assert (seq? []))
(assert (not (seq? nil)))
(assert (= 1 (first [1 2 3 4])))
(assert (= [2 3 4] (next [1 2 3 4])))
(assert (= nil (next [])))
(assert (= [1 2 3 4] (cons 1 [2 3 4])))
(assert (= [1 2 3 4] (conj [1 2] 3 4)))

; ; threading macros
(assert (= (-> 1 (cons [2 3 4])) [1 2 3 4]))
(assert (= (->>  1 (conj [2 3 4])) [2 3 4 1]))

; ; basic math operators
(assert (= 3 (+ 1 2)))
(assert (= 3 (- 5 2)))
(assert (= -5 (- 5)))
(assert (= 10 (* 5 2)))
(assert (= 5 (/ 10 2)))
(assert (= 0.50000 (/ 2)))
(assert (> 10 9 8 7 6 1 -1 -10))
(assert (< -10 1 2 3 4 10 23.32423432 100000))
(assert (>= 10 10 10 9 8 7 7 7 5))
(assert (<= -1.5 -1 0 0  0  0 0 0 1 2 3 4 5))

; ; type initialization functions
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
(assert (= 1 (int 1.5677)))
(assert (= 3.00000 (float 3)))

; ; type checking functions
(assert (int? 10))
(assert (not (int? 10.0)))
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
(assert (= 10 (-> 5 (+ 5) (int))))
(assert (= "hello" (-> 0 ["hello" 1 2 #{}])))

; ; simple function definition
(def dec (fn* [i] (int (- i 1))))
(assert (= 9 (dec 10)))

; ; simple recursive function with variadic args
(def down-range (fn* down-range [start & args]
    (if (> start 0)
        (cons start (down-range (dec start)))
        [0])))
(assert (= '(5 4 3 2 1 0) (down-range 5)))

; ; complex recursive function
(def reverse (fn* reverse [coll]
    (if (not (seq? coll))
        (throw "argument must be a sequence"))
    (if (nil? (next coll))
        [(first coll)]
        (let [first-value   (first coll)
               reversed      (reverse (next coll))]
            (conj reversed first-value)))))
(assert (= '(5 4 3 2 1) (reverse '(1 2 3 4 5))))

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

(def sum-through-let (let [numbers [1 2 3 4 5]]
                        (do (print "Numbers: " numbers)
                            (let [sum (apply-seq + numbers)]
                                (print "Sum of numbers: " sum)
                                sum))))
(assert (= (+ 1 2 3 4 5) sum-through-let))


(defn square [num] (* num num))
(defn positive? [num] (> num 0))

(assert (= '(0 1 2) (range 3)))
(assert (= '(4 5 6 7 8) (range 4 9)))
(assert (= '(4 6 8) (range 4 9 2)))

(assert (= [1 4 9 25] (map square [1 2 3 5])))
(assert (= '(1 4 9 25) (map square '(1 2 3 5))))

(assert (= [2 5 4 1] (filter positive? [-1 0 -2 2 5 4 1])))

(assert (= 45 (reduce + (range 10))))
(assert (= 55 (reduce + 10 (range 10))))
(assert (= 6 (reduce + [1 2 3])))
(assert (= 26 (reduce + 20 [1 2 3])))

