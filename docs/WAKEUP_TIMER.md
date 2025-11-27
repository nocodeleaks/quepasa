# Wake-Up Timer System

## Overview
The Wake-Up Timer system allows QuePasa to schedule daily WhatsApp presence activation at a specific hour. This is useful for simulating device activity at predetermined times without requiring manual interaction.

## Features
- Schedule presence activation at one specific hour each day
- Automatically sets presence to online at the configured hour
- Returns to offline after a configurable duration
- Uses server's local timezone
- Automatic daily scheduling (repeats every 24 hours)

## Configuration

### Environment Variables

1. **WAKEUP_HOUR** (optional)
   - Single hour of day when presence should be activated
   - Format: Integer from 0 to 23 (hour only, minutes always at :00)
   - Example: `WAKEUP_HOUR=9` (activates at 09:00)
   - Default: Empty (disabled)

2. **WAKEUP_DURATION** (optional)
   - Duration in seconds to keep presence online
   - Format: Integer (seconds)
   - Example: `WAKEUP_DURATION=30`
   - Default: 10 seconds

### Example Configurations

```bash
# Wake up at 9 AM for 10 seconds (default duration)
WAKEUP_HOUR=9

# Wake up at 2 PM for 30 seconds
WAKEUP_HOUR=14
WAKEUP_DURATION=30

# Wake up at midnight for 5 seconds
WAKEUP_HOUR=0
WAKEUP_DURATION=5
```

## How It Works

1. **Initialization**: When QuePasa starts, the WakeUpScheduler is created for each connection
2. **Hour Parsing**: The `WAKEUP_HOUR` value is parsed and validated (must be 0-23)
3. **Scheduling**: A background goroutine checks every 5 minutes if current time matches the wake-up hour
4. **Activation**: When the hour arrives (within 10-minute window):
   - Presence is set to `available` (online)
   - A timer is started for the configured duration
5. **Deactivation**: After the duration expires:
   - Presence is set to `unavailable` (offline)
   - Next wake-up is scheduled for tomorrow at the same hour
6. **Daily Repetition**: The process repeats automatically every 24 hours

## Technical Details

### Components

- **WakeUpScheduler**: Main scheduler component
  - Location: `src/whatsmeow/whatsmeow_wakeup_scheduler.go`
  - Manages wake-up timing and execution
  - Runs in background goroutine

- **Configuration**: Environment settings
  - Location: `src/environment/whatsapp_settings.go`
  - Defines `WAKEUP_HOUR` and `WAKEUP_DURATION` variables

- **Integration**: WhatsmeowConnection
  - Location: `src/whatsmeow/whatsmeow_connection.go`
  - Embeds WakeUpScheduler in connection lifecycle

### Timing Behavior

- Scheduler checks every 5 minutes (optimized for once-per-day execution)
- Triggers when current time is within 10 minutes of configured hour
- Only executes once per day
- Automatically reschedules for next day after execution
- Times are based on server's local timezone

## Examples by Timezone

### Brazil (UTC-3)
```bash
# Wake up at 9:00 AM Brazil time
WAKEUP_HOUR=9

# Wake up at 6:00 PM Brazil time  
WAKEUP_HOUR=18
```

### Portugal (UTC+0/UTC+1)
```bash
# Wake up at 8:00 AM Portugal time
WAKEUP_HOUR=8

# Wake up at 9:00 PM Portugal time
WAKEUP_HOUR=21
```

### UTC Server with Local Time Target
If your server is in UTC but you want wake-up at a specific local time:
```bash
# For 9 AM Brazil Time (UTC-3), set hour to 12 (9 + 3)
WAKEUP_HOUR=12

# For 9 AM Portugal Summer Time (UTC+1), set hour to 8 (9 - 1)
WAKEUP_HOUR=8
```

## Logging

The system logs the following events:

```
INFO  wake-up scheduler started for hour: 9:00
INFO  wake-up triggered at scheduled hour: 9:00
INFO  executing wake-up: setting presence to online for 10s
INFO  wake-up completed: presence set to offline after 10s
DEBUG next wake-up scheduled for: 2025-01-25 09:00
```

### Log Levels

- **INFO**: Scheduler start, wake-up triggers, presence changes
- **DEBUG**: Next wake-up scheduling details
- **WARN**: Configuration parsing errors, client unavailability

## Use Cases

### 1. Daily Check-in
Set a single daily presence activation to show account is active:
```bash
WAKEUP_HOUR=9
WAKEUP_DURATION=10
```

### 2. Business Hours Marker
Activate presence at start of business day:
```bash
WAKEUP_HOUR=8
WAKEUP_DURATION=30
PRESENCE=unavailable  # Stay offline otherwise
```

### 3. Evening Activity
Show presence during evening hours:
```bash
WAKEUP_HOUR=19
WAKEUP_DURATION=15
```

### 4. Midnight Maintenance
Wake up at midnight for system sync:
```bash
WAKEUP_HOUR=0
WAKEUP_DURATION=60
```

## Interaction with Other Features

### With Global PRESENCE Setting
```bash
WAKEUP_HOUR=9
WAKEUP_DURATION=30
PRESENCE=unavailable  # Offline by default, online during wake-up only
```

### Standalone Usage
```bash
WAKEUP_HOUR=14
WAKEUP_DURATION=20
# No PRESENCE variable - device manages its own status otherwise
```

## Troubleshooting

### Wake-up not triggering

1. **Check logs**: Look for scheduler initialization messages
2. **Verify hour**: Must be integer 0-23
3. **Check timezone**: Times use server's local timezone
4. **Wait for window**: Trigger happens within 10 minutes of hour start (scheduler checks every 5 minutes)

### Configuration errors

```
WARN failed to parse WAKEUP_HOUR '25': hour '25' out of range (must be 0-23)
WARN failed to parse WAKEUP_HOUR 'abc': invalid hour format 'abc'
```

### Client not available

```
WARN cannot send presence: client not available
DEBUG cannot send presence: push name not set
```

This happens when WhatsApp client is not yet connected or authenticated.

## Disabling the Feature

To disable wake-up timer, remove or comment out the variable:
```bash
# WAKEUP_HOUR=9
```

Or set it to empty:
```bash
WAKEUP_HOUR=
```

## Migration from AUTO_PRESENCE

If you were using the deprecated AUTO_PRESENCE feature:

**Old configuration:**
```bash
AUTO_PRESENCE=true
AUTO_PRESENCE_DELAY=30
```

**New equivalent (wake-up at 9 AM):**
```bash
WAKEUP_HOUR=9
WAKEUP_DURATION=30
```

The AUTO_PRESENCE feature has been completely removed in favor of this scheduled approach.

## Performance Considerations

- **Memory**: Minimal overhead (one goroutine per connection)
- **CPU**: Negligible (5-minute check interval, optimized for daily execution)
- **Network**: Brief presence update only during wake-up
- **Scaling**: Handles multiple connections independently
- **Efficiency**: Optimized for once-per-day execution pattern

## Security Notes

- Uses WhatsApp official presence protocol
- No credentials stored or transmitted
- Times based on server's secure time source
- Presence updates are standard WhatsApp operations

## Future Enhancements

Potential future improvements:
- Web UI for hour configuration
- Per-connection hour settings
- Timezone-aware scheduling
- Wake-up history tracking

## References

- WhatsApp Types: `go.mau.fi/whatsmeow/types`
- Presence Types: `PresenceAvailable`, `PresenceUnavailable`
- Scheduler Pattern: Background goroutine with context cancellation
- Environment Configuration: `src/environment/whatsapp_settings.go`

## Support

For issues or questions:
1. Check logs for error messages
2. Verify environment variable syntax
3. Consult QuePasa documentation
4. Open GitHub issue with logs and configuration
