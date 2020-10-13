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


; sequence operations -------------------------------
(defn seq? [arg] (impl? arg types/Seq))

(defn first [coll]
    (if (nil? coll)
        nil
        (if (not (seq? coll))
            (throw "argument must be a collection, not " (type coll))
            (coll.First))))

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
    (if (not (seq? coll))
        (throw "argument must be a collection, not " (type coll)))
    (if (nil? (coll.Next))
      '()
      (coll.Next)))


(defn cons [v coll]
    (if (not (seq? coll))
        (throw "argument must be a collection, not " (type coll)))
    (coll.Cons v))

(defn conj [coll & vals]
    (if (not (seq? coll))
        (throw "argument must be a collection, not " (type coll)))
    (apply coll.Conj vals))

; (defn swap! [atom f]
;   (atom.UpdateState f))

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
    (let [v   (first coll)
           rem (next coll)]
        (if (nil? rem)
            v
            (last (next coll)))))

(defn number? [num]
    (if (float? num)
        true
        (int? num)))

(defn even? [num]
    (= (mod num 2) 0.0))

(defn odd? [num]
    (= (mod num 2) 1.0))

(defn reverse [coll]
  (when-not (seq? coll)
    (throw "argument must be Seq"))
  (loop [result '() target coll]
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


(defn take
  ([n coll]
   (take n coll '()))
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

(defn map [f coll]
  (let [z '()]
    (doseq [x coll]
      (unsafe/swap z (conj z (f x))))
    z))

(defn map-indexed [f coll]
  (let [z '() i 0]
    (doseq [x coll]
      (unsafe/swap z (conj z (f x i)))
      (unsafe/swap i (inc i)))
    z))


(defn filter [f coll]
  (let [z '()]
    (doseq [x coll]
      (if (f x)
        (unsafe/swap z (conj z x))))
    z))


(defn filter-indexed [f coll]
  (let [z '() i 0]
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
   (apply coll1.Conj coll2))
  ([coll1 coll2 & more]
   (reduce concat (concat coll1 coll2) more)))


(defn realized? [x]
  (if (fn? x)
    false
    (realized* x)))


; source a file. Beware of circular dependency
(defn source [filename]
  (let [target (str filename ".lisp")
        content (read-file target)]
    (eval-string content)))


; important macros -----------------------------------


(defmacro apply [callable args]
    `(eval (cons ~callable ~args)))

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

(defmacro deref [symbol]
  `(deref* '~symbol ~symbol))


(defmacro future [& body]
  (let [body (cons 'do body)]
    `(future* ~body)))

(defmacro delay [& body]
  (let [body (cons 'do body)]
    `(fn [] (future* ~body))))

(defmacro force [x]
  `(deref* '~x (~x)))


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

; Type initialization functions ---------------------
(defn set [coll] (apply (type #{}) coll))
(defn list [& coll] (apply (type ()) coll))
(defn vector [& coll] (apply (type []) coll))
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
