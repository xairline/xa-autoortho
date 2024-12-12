use crate::misc::get_system_path;
use crate::plugin_debugln;
use ini::Ini;
use std::fs;
use std::fs::OpenOptions;
use std::io::{self, BufRead};
use std::path::Path;
use std::process::{Command, Stdio};
use std::str;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use std::thread;
use std::time::Duration;
pub fn mount(load_flight_on_start: bool) {
    let mounts = get_mounts();
    plugin_debugln!("Mounts: {:?}", mounts);
    let cmd_start_count = Arc::new(AtomicUsize::new(0));
    let wg = Arc::new(std::sync::Barrier::new(mounts.len()));
    let number_of_mounts = mounts.len();
    let mounts2 = mounts.clone();

    for mount in mounts {
        let cmd_start_count = Arc::clone(&cmd_start_count);
        let wg = Arc::clone(&wg);
        let home_dir = match dirs::home_dir() {
            Some(path) => path.to_string_lossy().into_owned(),
            None => {
                eprintln!("Failed to get home directory");
                return;
            }
        };
        let system_path = get_system_path();
        let plugin_path = Path::new(&system_path)
            .join("Resources")
            .join("plugins")
            .join("XA-autoortho");

        thread::spawn(move || {
            let parts: Vec<&str> = mount.split('|').collect();
            if parts.len() != 2 {
                eprintln!("Invalid mount format: {}", mount);
                return;
            }

            let root = parts[0];
            let mount_point = parts[1];

            let poison_file = Path::new(root).join(".poison");
            let log_file_path = Path::new(&home_dir).join(".autoortho-data/logs/autoortho.log");

            // Open log file for writing
            let log_file = OpenOptions::new()
                .append(true)
                .create(true)
                .write(true)
                .open(&log_file_path);

            if let Err(err) = log_file {
                eprintln!("Failed to open log file: {}", err);
                wg.wait();
                return;
            }

            // Remove the poison file if it exists
            if std::fs::remove_file(&poison_file).is_err() {
                // Poison file may not exist; ignoring error
            }

            // Start the command
            let command_path = plugin_path.join("autoortho_fuse");
            plugin_debugln!("Command Path: {:?}", command_path);
            let cmd = Command::new(command_path)
                .arg(root)
                .arg(mount_point)
                .stdout(Stdio::piped())
                .stderr(Stdio::piped())
                .spawn();
            plugin_debugln!("Command: {:?}", cmd);
            plugin_debugln!("Root: {:?}", root);
            plugin_debugln!("Mount Point: {:?}", mount_point);

            if let Err(err) = cmd {
                plugin_debugln!("Error starting command: {}", err);
                wg.wait();
                return;
            }

            let mut cmd = cmd.unwrap();
            cmd_start_count.fetch_add(1, Ordering::SeqCst);

            // Wait until the mount is ready
            loop {
                match is_fuse_mount_point(mount_point) {
                    Ok(true) => {
                        plugin_debugln!("Autoortho service is ready: {}", mount_point);
                        break;
                    }
                    Ok(false) => {
                        plugin_debugln!("Autoortho service is not ready: {}", mount_point);
                    }
                    Err(err) => {
                        plugin_debugln!("Error checking mount point: {}", err);
                    }
                }
                thread::sleep(Duration::from_secs(1));
            }

            // Notify that this thread is done initializing the mount
            wg.wait();

            // Handle cleanup upon termination
            let _ = cmd.wait().map(|status| {
                if status.success() {
                    plugin_debugln!("Command finished successfully");
                } else {
                    plugin_debugln!("Command finished with error");
                }
            });
        });
    }

    // Wait for all mounts to be ready if `load_flight_on_start` is true
    if load_flight_on_start {
        eprintln!("Locked, waiting for AO ready");
        loop {
            let mut fuse_mount_count = 0;
            let mounts2 = mounts2.clone();
            for mount in mounts2 {
                let parts: Vec<&str> = mount.split('|').collect();
                if parts.len() != 2 {
                    eprintln!("Invalid mount format: {}", mount);
                    return;
                }

                let mount_point = parts[1];
                match is_fuse_mount_point(mount_point) {
                    Ok(true) => {
                        fuse_mount_count += 1;
                    }
                    Ok(false) => {
                        eprintln!("Autoortho service is not ready: {}", mount_point);
                    }
                    Err(err) => {
                        eprintln!("Error checking mount point: {}", err);
                    }
                }
            }
            if fuse_mount_count == number_of_mounts {
                break;
            }
        }
    }
}

fn get_mounts() -> Vec<String> {
    let mut res = Vec::new();

    // Get the user's home directory
    let home_dir = match dirs::home_dir() {
        Some(path) => path,
        None => {
            eprintln!("Failed to get home directory");
            return res;
        }
    };

    // Load the configuration file
    let config_path = home_dir.join(".autoortho");
    let config = match Ini::load_from_file(&config_path) {
        Ok(cfg) => cfg,
        Err(err) => {
            eprintln!("Failed to read file: {}", err);
            return res;
        }
    };

    // Get the "paths" section
    let section = match config.section(Some("paths")) {
        Some(sec) => sec,
        None => {
            eprintln!("Failed to get section 'paths'");
            return res;
        }
    };

    // Retrieve the paths
    let xplane_path = section.get("xplane_path").unwrap_or_default();
    let scenery_path = section.get("scenery_path").unwrap_or_default();
    plugin_debugln!("XPlanePath: {}, SceneryPath: {}", xplane_path, scenery_path);

    // Get the folders under the specified path
    let scenery_dir = Path::new(scenery_path).join("z_autoortho/scenery");
    match fs::read_dir(&scenery_dir) {
        Ok(folders) => {
            for folder in folders {
                if let Ok(entry) = folder {
                    if entry.file_type().map(|ft| ft.is_dir()).unwrap_or(false) {
                        let region_name = entry.file_name();
                        let root = Path::new(scenery_path)
                            .join("z_autoortho/scenery")
                            .join(&region_name);
                        let mount = Path::new(xplane_path)
                            .join("Custom Scenery")
                            .join(&region_name);
                        res.push(format!("{}|{}", root.display(), mount.display()));
                    }
                }
            }
        }
        Err(err) => {
            eprintln!(
                "Failed to read directory '{}': {}",
                scenery_dir.display(),
                err
            );
        }
    }

    res
}

fn is_fuse_mount_point(path: &str) -> Result<bool, io::Error> {
    // Execute the `mount` command
    let output = Command::new("mount").output()?.stdout;

    // Read the output line by line
    let reader = io::BufReader::new(&*output);
    for line in reader.lines() {
        let line = line?;
        // Check if the line contains the desired path
        if line.contains(&format!(" on {} ", path)) {
            // Extract the filesystem type from the line
            if let Some(start) = line.find('(') {
                if let Some(end) = line.find(')') {
                    if end > start {
                        let fs_info = &line[start + 1..end];
                        // Split the filesystem information into fields
                        let fields: Vec<&str> = fs_info.split(',').map(|s| s.trim()).collect();
                        if fields.iter().any(|&fs| {
                            fs.starts_with("fuse")
                                || fs.contains("osxfuse")
                                || fs.contains("macfuse")
                        }) {
                            return Ok(true);
                        }
                    }
                }
            }
        }
    }

    Ok(false)
}
