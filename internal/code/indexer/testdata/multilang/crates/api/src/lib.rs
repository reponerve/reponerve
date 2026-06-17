use serde::Serialize;

pub struct Service {
    name: String,
}

pub fn run() -> String {
    "running".to_string()
}

impl Service {
    pub fn name(&self) -> &str {
        &self.name
    }
}

pub trait Store {
    fn save(&self, key: &str);
}
