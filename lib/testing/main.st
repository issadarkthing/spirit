
(def test-count 0)
(def success-count 0)
(def fail-count 0)

(defmacro test-block [& body]
  (let [body (cons 'do body)]
    `(do
       ~body
       (print)
       (time)
       (printf "Tests: %v tests %v success %v failed\n" test-count success-count fail-count))))

(defmacro test [desc & body]
  (def test-count (inc test-count))
  (let [body (cons 'do body)]
   `(try
      (do
        ~body
        (def success-count (inc success-count))
        (printf "[PASS] %v\n" ~desc))
      (fn [err]
        (do
          (def fail-count (inc fail-count))
          (printf "[FAIL] %v\n%v\n" ~desc err))))))
    
