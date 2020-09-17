
(def list (tview/new-list))
(def app (tview/new-app))

(tview/list-add-item list "List item 1" "sample" \a nil)
(tview/list-add-item list "List item 1" "sample" \b nil)
(tview/list-add-item list "List item 1" "sample" \c nil)
(tview/list-add-item list "List item 1" "sample" \d nil)
(tview/list-add-item list "Quit" "Press to exit" \q (fn []
                                                      (app.Stop)))

(defn next-list [l]
  (l.SetCurrentItem (let [curr-index (l.GetCurrentItem)]
                      (if (= curr-index (dec (l.GetItemCount)))
                        0
                        (inc curr-index)))))

(defn prev-list [l]
  (l.SetCurrentItem (dec (l.GetCurrentItem))))

(list.SetBackgroundColor tview/default-color)
(list.SetBorder true)

(tview/app-set-before-draw 
  app 
  (fn [screen]
    (screen.Clear)
    false))

(tview/app-set-input-capture 
  app 
  (fn [e]
    (let [ch (to-type types/Char (e.Rune))]
      (case ch
        \j (next-list list)
        \k (prev-list list)
        nil))))

(app.SetRoot list true)
(app.EnableMouse true)
(app.Run)

 

