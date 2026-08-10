local fixed = require("darkmagic.ui.fixed_panel")
return fixed.overlay({
    sheet="data/global/ui/PANEL/supertransmogrifier.dc6", x=80, y=64, close_x=275, close_y=15,
    close_label="darkmagic.cube.close",
    labels={{key="darkmagic.overlay.state_unavailable",x=160,y=370,width=250}},
})
