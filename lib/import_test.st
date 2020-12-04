(import "./subdir/content.st")
(import "./subdir/content-2.st")

(assert "with namespace" (= 1000 content/x))
(assert "without namespace" (= 1010 z))
