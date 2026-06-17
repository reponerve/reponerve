const std = @import("std");

pub const Handler = struct {
    pub fn health() []const u8 {
        return "ok";
    }
};

pub fn bootstrap() []const u8 {
    return "ready";
}
