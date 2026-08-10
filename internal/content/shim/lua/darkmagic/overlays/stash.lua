local fixed = require("darkmagic.ui.fixed_panel")
return fixed.overlay({
    sheet="data/global/ui/PANEL/TradeStash.DC6", x=80, y=64, close_x=272, close_y=15,
    close_label="darkmagic.stash.close",
    labels={{key="darkmagic.overlay.state_unavailable",x=160,y=390,width=250}},
})
