package whatsmeow

// Example of how to use the new GroupManager pattern

/*
Usage Example:

// Before refactoring (old way):
connection := &WhatsmeowConnection{...}
link, err := connection.GetInvite(groupId)
groups, err := connection.GetJoinedGroups()

// After refactoring (new way):
connection := &WhatsmeowConnection{...}
groupManager := connection.GetGroupManager()

// All group operations are now done through the manager
link, err := groupManager.GetInvite(groupId)
groups, err := groupManager.GetJoinedGroups()
groupInfo, err := groupManager.GetGroupInfo(groupId)
newGroup, err := groupManager.CreateGroup("My Group", []string{"5511999999999@s.whatsapp.net"})

// Benefits:
// 1. Clear separation of concerns
// 2. WhatsmeowConnection is smaller and more focused
// 3. Group operations are centralized in a dedicated manager
// 4. Interface-based design allows for easy testing and mocking
// 5. Lazy initialization of GroupManager only when needed
*/
