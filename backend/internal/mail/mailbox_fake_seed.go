package mail

// PutFolder safely seeds one account-scoped folder by value for deterministic fake scenarios.
func (f *FakeMailEngine) PutFolder(folder EngineFolder) bool {
	if f == nil || !validAccountID(folder.AccountID) || !validOpaqueID(folder.FolderID) || !validOpaqueIDOrEmpty(folder.ParentID) || !validSafeText(folder.Name, MaxEngineMIMEFieldBytes, true) || !validFakeFolderRole(folder.Role) {
		return false
	}
	nameKey, err := normalizeFolderName(folder.Name)
	if err != nil {
		return false
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.accounts[folder.AccountID]; !exists {
		return false
	}
	folders := f.folders[folder.AccountID]
	if existing, exists := folders[folder.FolderID]; exists {
		return existing == folder
	}
	if len(folders) >= MaxEngineFolderItems || folder.Role != EngineFolderRoleCustom && folder.ParentID != "" {
		return false
	}
	if folder.ParentID != "" {
		if _, exists := folders[folder.ParentID]; !exists {
			return false
		}
	}
	for _, existing := range folders {
		if folder.Role != EngineFolderRoleCustom && existing.Role == folder.Role {
			return false
		}
		if folder.Role != EngineFolderRoleCustom || existing.Role != EngineFolderRoleCustom || existing.ParentID != folder.ParentID {
			continue
		}
		existingNameKey, err := normalizeFolderName(existing.Name)
		if err != nil || existingNameKey == nameKey {
			return false
		}
	}
	if err := f.reserveOpaqueIDsLocked(folder.FolderID); err != nil {
		return false
	}
	if folders == nil {
		folders = make(map[string]EngineFolder)
		f.folders[folder.AccountID] = folders
	}
	folders[folder.FolderID] = folder
	return true
}

func validFakeFolderRole(role EngineFolderRole) bool {
	switch role {
	case EngineFolderRoleCustom,
		EngineFolderRoleInbox,
		EngineFolderRoleSent,
		EngineFolderRoleDrafts,
		EngineFolderRoleArchive,
		EngineFolderRoleTrash,
		EngineFolderRoleJunk:
		return true
	default:
		return false
	}
}
