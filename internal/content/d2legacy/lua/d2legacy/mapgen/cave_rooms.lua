-- Act I cave room selection from authored W/E/S/N connection masks.

local M = {}

local WEST, EAST, SOUTH, NORTH = 1, 2, 4, 8

local function cell_key(cell)
	local x, y = cell.x, cell.y
	if x == 0 then x = 0 end
	if y == 0 then y = 0 end
	return x .. ":" .. y
end

local function add(masks, cell, bit)
    local name = cell_key(cell)
    masks[name] = (masks[name] or 0) + bit
end

function M.connection_masks(links)
    local masks = {}
    for _, link in ipairs(links) do
        local a, b = {x = link.ax, y = link.ay}, {x = link.bx, y = link.by}
        if b.x == a.x - 1 then add(masks, a, WEST); add(masks, b, EAST)
        elseif b.x == a.x + 1 then add(masks, a, EAST); add(masks, b, WEST)
        elseif b.y == a.y + 1 then add(masks, a, SOUTH); add(masks, b, NORTH)
        elseif b.y == a.y - 1 then add(masks, a, NORTH); add(masks, b, SOUTH)
        end
    end
    return masks
end

-- Direct lookup avoids making callers depend on the internal coordinate-key
-- representation. It is also easier to read at the room-assembly call site.
function M.connection_mask(cell, links)
	local mask = 0
	for _, link in ipairs(links) do
		local other
		if link.ax == cell.x and link.ay == cell.y then other = {x = link.bx, y = link.by} end
		if link.bx == cell.x and link.by == cell.y then other = {x = link.ax, y = link.ay} end
		if other then
			if other.x == cell.x - 1 then mask = mask + WEST
			elseif other.x == cell.x + 1 then mask = mask + EAST
			elseif other.y == cell.y + 1 then mask = mask + SOUTH
			elseif other.y == cell.y - 1 then mask = mask + NORTH
			end
		end
	end
	return mask
end

function M.preset_definition(role, mask)
    if role == nil or role == "" then return 52 + mask end
    local ordinal = ({[WEST] = 0, [EAST] = 1, [SOUTH] = 2, [NORTH] = 3})[mask]
    assert(ordinal ~= nil, "d2legacy mapgen: special cave room must have one door")
    if role == "previous-level" then return 83 + ordinal end
    if role == "next-level" then return 87 + ordinal end
    error("d2legacy mapgen: unknown special cave role " .. tostring(role))
end

return M
