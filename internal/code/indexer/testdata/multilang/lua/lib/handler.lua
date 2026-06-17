local json = require("json")

local Handler = {}
Handler.__index = Handler

function Handler.new()
    return setmetatable({}, Handler)
end

function Handler:health()
    return "ok"
end

local function bootstrap()
    return "ready"
end

return Handler
