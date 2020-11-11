(ns 'types)

(def Int    (type 0))
(def Float  (type 0.0))
(def Vector (type []))
(def List   (type ()))
(def Set    (type #{}))
(def Bool   (type true))
(def String (type "specimen"))
(def Keyword(type :specimen))
(def Symbol (type 'specimen))
(def Char   (type \a))
(def Fn     (type (fn* [])))
(def HashMap(type {}))
(def Future (type (future* 1)))
(def LazySeq (type (lazy-range* 0 0 1)))

(ns 'core)


(def fn (macro* fn [& decl]
         (decl.Cons 'fn*)))

(def defn (macro* defn [name & fdecl]
           (let [with-name (fdecl.Cons name)
                  func      (with-name.Cons 'fn*)]
               `(def ~name ~func))))

(def defmacro (macro* defmacro [name & mdecl]
               (let [with-name (mdecl.Cons name)
                      macro     (with-name.Cons 'macro*)]
                   `(def ~name ~macro))))

(defn nil? [arg] (= nil arg))

; io operations -----------------------------------

(defn read
  ([]
   (read* ""))
  ([prompt]
   (read* prompt)))

(defmacro tap [expr]
  `(let [result ~expr]
     (print result)
     result))
; math operations -------------------------------

(defn positive? [x]
  (> x 0))


; sequence operations -------------------------------
(defn seq? [arg] (impl? arg types/Seq))

(defn lazy-range
  ([] (lazy-range* 0 922337203685477 1))
  ([max] (lazy-range* 0 max 1))
  ([min max] (lazy-range* min max 1))
  ([min max step] (lazy-range* min max step)))

(defn range
  ([max]
   (range 0 max 1))
  ([min max]
   (range min max 1))
  ([min max step]
   (loop [i min res '()]
     (if (>= i max)
       res
       (recur (+ i step) (conj res i))))))

(defn range-vec
  ([max]
   (range-vec 0 max 1))
  ([min max]
   (range-vec min max 1))
  ([min max step]
   (loop [coll []
          index min]
     (if (= index max)
       coll
       (recur (conj coll index) (+ step index))))))

(defn first [coll]
  (if (not (seq? coll))
    (throw "argument must be a collection, not " (type coll))
    (coll.First)))

(defn second [coll]
    (first (next coll)))

(defn third [coll]
  (first (next (next coll))))

(defn next [coll]
    (if (not (seq? coll))
        (throw "argument must be a collection, not " (type coll)))
    (coll.Next))

; same as next but returns empty list if no next member instead of nil
(defn rest [coll]
  (if (> (count coll) 1)
    (next coll)
   (cond 
     (or (lazy-seq? coll) (list? coll)) '()
     (vector? coll) [])))

(defn cons [v coll]
    (if (not (seq? coll))
        (throw "argument must be a collection, not " (type coll)))
    (coll.Cons v))


(defn conj-1 [coll v]
  (coll.Conj v))

(defn conj [coll & others]
  (reduce conj-1 coll others))
  
; (defn swap! [atom f]
;   (atom.UpdateState f))

(defn drop [n coll]
  (if (zero? n)
    coll
    (recur (dec n) (next coll))))

(defn into [to from]
  (if (empty? from)
    to
    (recur (conj to (first from)) (next from))))

(defn empty? [coll]
    (if (nil? coll)
        true
        (nil? (first coll))))

(defn cons [val coll]
    (if (nil? coll)
        (cons val ())
        (if (seq? coll)
            (coll.Cons val)
            (throw "cons cannot be done for " (type coll)))))

(defn last [coll]
  (if (vector? coll)
    (coll (- (count coll) 1))
    (let [v   (first coll)
          rem (next coll)]
      (if (nil? rem)
        v
        (last (next coll))))))

(defn number? [num]
    (if (float? num)
        true
        (int? num)))

(defn some [f coll]
  (if (or (empty? coll) (nil? coll))
    false
    (or (f (first coll)) (some f (next coll)))))  

(defn every [f coll]
  (if (or (empty? coll) (nil? coll))
    true
    (and (f (first coll)) (every f (next coll)))))  

(defn find [f coll]
  (when (and coll (not (empty? coll)))
    (let [v (first coll)]
      (if (f v)
        v
        (recur f (next coll))))))

(defn even? [num]
    (= (mod num 2) 0))

(defn odd? [num]
    (= (mod num 2) 1))

(defn reverse [coll]
  (when-not (seq? coll)
    (throw "argument must be Seq"))
  (loop [result (empty coll) target coll]
    (if (empty? target)
      result
      (recur (cons (first target) result) (rest target)))))

(defn inc [num]
    (if (not (number? num))
        (throw "argument must be a number"))
    (if (int? num)
        (int (+ 1 num))
        (+ 1 num)))

(defn zero? [num]
  (= num 0))

(defn dec [num]
    (if (not (number? num))
        (throw "argument must be a number"))
    (if (int? num)
        (int (- num 1))
        (- num 1)))

(defn count [coll]
    (if (not (seq? coll))
      (throw "argument must be a Seq"))
    (coll.Size))

(defn empty [coll]
  (if (not (seq? coll))
    (throw "argument must be a Seq"))
  (cond 
    (string? coll) ""
    (list? coll) '() 
    (vector? coll) []
    (lazy-seq? coll) '()))
    

(defn take
  ([n coll]
   (take n coll (empty coll)))
  ([n coll acc]
   (if (= n (count acc))
     acc
     (recur n (next coll) (conj acc (first coll))))))

(defn reduce
  ([f coll]
   (reduce f (first coll) (next coll)))
  ([f acc coll]
   (let [z acc]
     (doseq [x coll]
       (unsafe/swap z (f z x)))
     z)))

(defn reduce-indexed
  ([f coll]
   (reduce-indexed f (first coll) (next coll)))
  ([f acc coll]
   (let [z acc i 0]
     (doseq [x coll]
       (unsafe/swap z (f z x i))
       (unsafe/swap i (inc i)))
     z)))

(defn map
  ([f coll]
   (if (empty? coll)
       coll
       (cons (f (first coll)) (map f (rest coll)))))
  ([f c1 c2]
   (let [s1 c1 s2 c2]
     (if (or (empty? s1) (empty? s2))
         (empty s1)
         (cons (f (first s1) (first s2))
               (map f (rest s1) (rest s2))))))
  ([f c1 c2 c3]
   (let [s1 c1 s2 c2 s3 c3]
     (if (or  (empty? s1) (empty? s2) (empty? s3))
         (empty s1)
         (cons (f (first s1) (first s2) (first s3))
               (map f (rest s1) (rest s2) (rest s3)))))))

(defn map-indexed
  ([f coll]
   (map-indexed f coll (empty coll) 0))
  ([f coll acc i]
   (let [x (first coll)
         xs (rest coll)]
     (if (nil? x)
       acc
       (recur f xs (conj acc (f x i)) (inc i))))))


(defn filter [f coll]
  (let [z (empty coll)]
    (doseq [x coll]
      (if (f x)
        (unsafe/swap z (conj z x))))
    z))


(defn filter-indexed [f coll]
  (let [z (empty coll) i 0]
    (doseq [x coll]
      (if (f x i)
        (unsafe/swap z (conj z x)))
      (unsafe/swap i (inc i)))
    z))

; deref all
(defn deref-all [& coll]
  (doseq [v coll]
    (deref v)))

(defn concat 
  ([coll1 coll2]
   (<> coll1.Conj coll2))
  ([coll1 coll2 & more]
   (reduce concat (concat coll1 coll2) more)))


(defn realized? [x]
  (if (fn? x)
    false
    (realized* x)))
 
(defn flat 
  ([coll]
   (flat coll (empty coll)))
  ([coll empty-seq]
   (if (nil? coll) 
     empty-seq
     (let [x (first coll)
           xs (next coll)]
       (if (seq? x)
         (concat (flat x empty-seq) (flat xs empty-seq))
         (cons x (flat xs empty-seq)))))))

(defn assoc [coll & args]
  (if (vector? coll)
    (coll.Assoc (first args) (second args))
    (<> assoc* coll args)))

(defn abs [x]
  (if (< x 0) (- x) x))

; wait for any of the channels realized and returns the value
(defn wait-any [forms]
  (if (some realized? forms)
    (find realized? forms)
    (recur forms)))
  
; wait for all of the channels realized and returns Seq of values
; wait-all does not wait sequentially instead it waits for whichever comes first
; and continue wait for others
(defn wait-all [forms]
  (if (every (fn [x] (not (future? x))) forms)
    forms
    (recur (map 
             (fn [v] 
               (if (future? v) 
                 (if (realized? v) (deref v) v)
                 v)) 
             forms))))

; source a file. Beware of circular dependency
(defn source [filename]
  (def *file* filename)
  (eval-string (read-file filename)))

(defn complement
  [f] 
  (fn 
    ([] (not (f)))
    ([x] (not (f x)))
    ([x y] (not (f x y)))
    ([x y & zs] (not (<> f x y zs)))))


(defn take-last [n coll]
  (drop (- (count coll) n) coll))

(defn sum [coll]
  (<> + coll))


(defn constantly [x]
  (fn [& args] x))
                     
(defn identity [x] x)
  
(defn not=
  ([x] false)
  ([x y] (not (= x y)))
  ([x y & more]
   (not (<> = x y more))))

; important macros -----------------------------------

(defmacro -> [x & forms]
  (loop [x x forms forms]
    (if (empty? forms)
      x
      (let [form (first forms)
            threaded (if (list? form)
                       (let [f (first form)
                             args (next form)] 
                         `(<> ~f '~x '~args))
                       `(list '~form '~x))]
        (recur (eval threaded) (next forms))))))

(defmacro ->> [x & forms]
  (loop [x x forms forms]
    (if (empty? forms)
      x
      (let [form (first forms)
            threaded (if (list? form)
                       `(conj '~form '~x)
                       `(list '~form '~x))]
        (recur (eval threaded) (next forms))))))

(defmacro when [expr & body]
    (let [body (cons 'do body)]
     `(if ~expr ~body)))

(defmacro when-not [expr & body]
    (let [body (cons 'do body)]
     `(if (not ~expr) ~body)))

(defmacro assert
    ([expr] (let [message "assertion failed"]
                `(when-not ~expr (throw ~message))))
    ([expr message] `(when-not ~expr (throw ~message))))


(defmacro future [& body]
  (let [body (cons 'do body)]
    `(future* ~body)))

(defmacro delay [& body]
  (let [body (cons 'do body)]
    `(fn [] (future* ~body))))

(defmacro force [x]
  `(deref* '~x (~x)))

(defmacro cond [& clauses]
  (when clauses
    (when (not (even? (count clauses)))
      (throw "cond requires an even number of forms"))
    `(if (eval (first '~clauses))
       (eval (second '~clauses))
       (eval (cons 'cond (drop 2 '~clauses))))))

; Type check functions -------------------------------
(defn is-type? [typ arg] (= typ (type arg)))
(defn set? [arg] (is-type? #{} arg))
(defn list? [arg] (is-type? types/List arg))
(defn fn? [arg] (is-type? types/Fn arg))
(defn vector? [arg] (is-type? types/Vector arg))
(defn int? [arg] (is-type? types/Int arg))
(defn float? [arg] (is-type? types/Float arg))
(defn boolean? [arg] (is-type? types/Bool arg))
(defn string? [arg] (is-type? types/String arg))
(defn keyword? [arg] (is-type? types/Keyword arg))
(defn symbol? [arg] (is-type? types/Symbol arg))
(defn future? [arg] (is-type? types/Future arg))
(defn lazy-seq? [arg] (is-type? types/LazySeq arg))

; Type initialization functions ---------------------
(defn set [coll] (<> (type #{}) coll))
(defn list [& coll] (<> (type '()) coll))
(defn vector [& coll] (<> (type []) coll))
(defn int [arg] (to-type (type 0) arg))
(defn float [arg] (to-type (type 0.0) arg))
(defn boolean [arg] (true? arg))

; boolean operations --------------------------------
(defn true? [arg]
    (if (nil? arg)
        false
        (if (boolean? arg)
            arg
            true)))

(defn and [& args]
  (reduce (fn [acc x] (and* acc x)) args))


(defn or [& args]
  (reduce (fn [acc x] (or* acc x)) args))


(defn not [arg] (= false (true? arg)))

