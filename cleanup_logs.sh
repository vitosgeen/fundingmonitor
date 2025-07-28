#!/bin/bash

# Funding Monitor Log Cleanup Script
# This script helps clean up old log files to manage disk space

LOG_DIR="funding_logs"
DEFAULT_DAYS=7

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -d, --days DAYS     Number of days to keep (default: $DEFAULT_DAYS)"
    echo "  -l, --list          List all log files with their sizes"
    echo "  -s, --stats         Show statistics about log files"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Clean up logs older than 7 days"
    echo "  $0 -d 30             # Clean up logs older than 30 days"
    echo "  $0 -l                # List all log files"
    echo "  $0 -s                # Show log statistics"
}

# Function to list log files
list_logs() {
    if [ ! -d "$LOG_DIR" ]; then
        echo "Log directory '$LOG_DIR' does not exist."
        return
    fi
    
    echo "Log files in $LOG_DIR:"
    echo "========================"
    find "$LOG_DIR" -name "*.log" -type f -exec ls -lh {} \; | sort -k5 -hr
}

# Function to show statistics
show_stats() {
    if [ ! -d "$LOG_DIR" ]; then
        echo "Log directory '$LOG_DIR' does not exist."
        return
    fi
    
    echo "Log Statistics:"
    echo "==============="
    
    total_files=$(find "$LOG_DIR" -name "*.log" -type f | wc -l)
    total_size=$(find "$LOG_DIR" -name "*.log" -type f -exec du -ch {} + | tail -1 | cut -f1)
    
    echo "Total log files: $total_files"
    echo "Total size: $total_size"
    
    if [ $total_files -gt 0 ]; then
        echo ""
        echo "Files by date:"
        find "$LOG_DIR" -name "*.log" -type f | sed 's/.*\/\([0-9-]*\)\.log/\1/' | sort | uniq -c | sort -k2
    fi
}

# Function to clean up old logs
cleanup_logs() {
    local days=$1
    
    if [ ! -d "$LOG_DIR" ]; then
        echo "Log directory '$LOG_DIR' does not exist."
        return
    fi
    
    echo "Cleaning up log files older than $days days..."
    
    # Count files before cleanup
    files_before=$(find "$LOG_DIR" -name "*.log" -type f | wc -l)
    
    # Remove old files
    find "$LOG_DIR" -name "*.log" -type f -mtime +$days -delete
    
    # Count files after cleanup
    files_after=$(find "$LOG_DIR" -name "*.log" -type f | wc -l)
    files_removed=$((files_before - files_after))
    
    echo "Removed $files_removed log files."
    echo "Remaining: $files_after log files"
}

# Parse command line arguments
DAYS=$DEFAULT_DAYS
ACTION="cleanup"

while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--days)
            DAYS="$2"
            shift 2
            ;;
        -l|--list)
            ACTION="list"
            shift
            ;;
        -s|--stats)
            ACTION="stats"
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Execute the requested action
case $ACTION in
    "list")
        list_logs
        ;;
    "stats")
        show_stats
        ;;
    "cleanup")
        cleanup_logs $DAYS
        ;;
esac 