local fixed = require("darkmagic.ui.fixed_panel")
return fixed.overlay({
    sheet="data/global/ui/MENU/waygatebackground.dc6", x=80, y=64, close_x=272, close_y=14,
    close_label="darkmagic.waypoint.close",
    labels={{key="darkmagic.waypoint.unavailable",x=160,y=385,width=250}},
})
