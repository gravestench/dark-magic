local render=require("dm.render/v1"); local input=require("dm.input/v1"); local scenes=require("dm.scene/v1")
return {blocks_update_below=true,create=function(self) self.root=render.create("modal");self.root:set_position(400,300);self.root:fill_rect(360,260,12,10,9,250) end,update=function(self) if input.pressed("pause") or input.pressed("cancel") then scenes.pop() end end}
