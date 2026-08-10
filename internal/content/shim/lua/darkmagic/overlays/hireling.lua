local fixed = require("darkmagic.ui.fixed_panel")
return fixed.overlay({
    sheet="data/global/ui/PANEL/NpcInv.dc6", x=80, y=64, close_x=272, close_y=15,
    close_label="darkmagic.hireling.close",
    labels={{key="darkmagic.hireling.unavailable",x=160,y=205,width=250}},
})
