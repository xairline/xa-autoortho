// Define a custom macro that always includes the plugin name.
extern crate xplm;
#[macro_export]
macro_rules! plugin_debugln {
    ($($arg:tt)*) => {
        xplm::debugln!("[XA AutoOrtho Launcher] {}", format!($($arg)*))
    };
}
