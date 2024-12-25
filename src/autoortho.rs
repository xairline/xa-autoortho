use crate::misc::get_system_path;
use crate::plugin_debugln;
use ini::Ini;
use std::fs;
use std::fs::OpenOptions;
use std::io::{self, BufRead};
use std::path::Path;
use std::process::{Command, Stdio};
use std::str;

// --------------------------------------------------
// Main function that mounts all folders without
// blocking or monitoring the subprocesses
// --------------------------------------------------
pub fn mount(_load_flight_on_start: bool) {
    let mounts = get_mounts();
    plugin_debugln!("Mounts: {:?}", mounts);
    let mut number_of_mounts = 0;
    // Grab home directory
    let home_dir = match dirs::home_dir() {
        Some(path) => path.to_string_lossy().into_owned(),
        None => {
            eprintln!("Failed to get home directory");
            return;
        }
    };

    // Build path to your plugin folder
    let system_path = get_system_path();
    let plugin_path = Path::new(&system_path)
        .join("Resources")
        .join("plugins")
        .join("XA-autoortho");

    // For each mount, just spawn a subprocess and return
    for mount in mounts.iter() {
        number_of_mounts += 1;
        let parts: Vec<&str> = mount.split('|').collect();
        if parts.len() != 2 {
            eprintln!("Invalid mount format: {}", mount);
            continue;
        }

        let root = parts[0];
        let mount_point = parts[1];

        match is_fuse_mount_point(mount_point) {
            Ok(true) => {
                plugin_debugln!("Autoortho service is already mounted: {}", mount_point);
                continue;
            }
            Ok(false) => {
                eprintln!("Autoortho service is not mounted: {}", mount_point);
            }
            Err(err) => {
                eprintln!("Error checking mount point: {}", err);
            }
        }

        // Example: Poison file logic
        let poison_file = Path::new(root).join(".poison");
        if std::fs::remove_file(&poison_file).is_err() {
            // If .poison file doesn't exist, ignore
        }

        // Optional: open a log file, etc.
        let log_file_path = Path::new(&home_dir).join(".autoortho-data/logs/autoortho.log");
        if let Err(err) = OpenOptions::new()
            .append(true)
            .create(true)
            .write(true)
            .open(&log_file_path)
        {
            eprintln!("Failed to open log file: {}", err);
            // Continue anyway
        }

        // The "autoortho_fuse" binary path
        let command_path = plugin_path.join("autoortho_fuse");
        plugin_debugln!("Command Path: {:?}", command_path);
        plugin_debugln!("Root: {:?}", root);
        plugin_debugln!("Mount Point: {:?}", mount_point);

        // let umount = Command::new("umount")
        //     .arg(mount_point)
        //     .stdout(Stdio::null())
        //     .stderr(Stdio::null())
        //     .spawn();
        // return;

        // "Fire-and-forget" spawn:
        //   - stdout/stderr -> dev/null (or you can direct them to a file)
        //   - Do NOT call wait() => main thread is not blocked
        let child = Command::new(command_path)
            .arg(root)
            .arg(mount_point)
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .spawn();

        match child {
            Ok(_) => {
                plugin_debugln!("Spawned subprocess for mount successfully.");
            }
            Err(err) => {
                plugin_debugln!("Failed to spawn subprocess for mount: {}", err);
            }
        }
    }

    eprintln!("Locked, waiting for AO ready");
    loop {
        let mut fuse_mount_count = 0;
        for mount in mounts.iter() {
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

// --------------------------------------------------
// Reads [paths] from ~/.autoortho and returns each
// "root|mount" pair for autoortho usage
// --------------------------------------------------
fn get_mounts() -> Vec<String> {
    let mut res = Vec::new();

    // Get user’s home directory
    let home_dir = match dirs::home_dir() {
        Some(path) => path,
        None => {
            eprintln!("Failed to get home directory");
            return res;
        }
    };

    // Load config file
    let config_path = home_dir.join(".autoortho");
    let config = match Ini::load_from_file(&config_path) {
        Ok(cfg) => cfg,
        Err(err) => {
            eprintln!("Failed to read file: {}", err);
            return res;
        }
    };

    // Section: paths
    let section = match config.section(Some("paths")) {
        Some(sec) => sec,
        None => {
            eprintln!("Failed to get section 'paths'");
            return res;
        }
    };

    // Retrieve the relevant paths
    let xplane_path = section.get("xplane_path").unwrap_or_default();
    let scenery_path = section.get("scenery_path").unwrap_or_default();
    plugin_debugln!("XPlanePath: {}, SceneryPath: {}", xplane_path, scenery_path);

    // Gather all subfolders under z_autoortho/scenery
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

                        // We'll store them in "root|mount" format
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

// --------------------------------------------------
// Utility that checks if the given path is a fuse mount
// (Not used in the new "no blocking" approach, but kept
//  for reference or if you need it in the future.)
// --------------------------------------------------
fn is_fuse_mount_point(path: &str) -> Result<bool, io::Error> {
    // Run the "mount" command
    let output = Command::new("mount").output()?.stdout;

    // Read line by line
    let reader = io::BufReader::new(&*output);
    for line in reader.lines() {
        let line = line?;
        // Check if the line contains something like "... on /path/to/mount ..."
        if line.contains(&format!(" on {} ", path)) {
            // Extract the filesystem type from the line
            if let Some(start) = line.find('(') {
                if let Some(end) = line.find(')') {
                    if end > start {
                        let fs_info = &line[start + 1..end];
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
