
(def list (tview/new-list))
(def app (tview/new-app))

(def empty (fn [] ()))

(tview/list-add-item list "List item 1" "sample" 0 empty)
(tview/list-add-item list "List item 1" "sample" 0 empty)
(tview/list-add-item list "List item 1" "sample" 0 empty)
(tview/list-add-item list "List item 1" "sample" 0 empty)
(tview/list-add-item list "Quit" "Press to exit" 0 (fn []
                                                      (app.Stop)))

(defn next-list [l]
  (l.SetCurrentItem (let [curr-index (l.GetCurrentItem)]
                      (if (= curr-index (dec (l.GetItemCount)))
                        0
                        (inc curr-index)))))

(defn prev-list [l]
  (l.SetCurrentItem (dec (l.GetCurrentItem))))

(list.ShowSecondaryText false)
(list.SetBackgroundColor tview/color-default)
(list.SetBorder true)
(list.SetHighlightFullLine true)
(list.SetSelectedTextColor tview/color-red)
(list.SetSelectedBackgroundColor tview/color-green)

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
        nil)
      e)))

(app.SetRoot list true)
(app.EnableMouse true)
(app.Run)

 

