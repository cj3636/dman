package config

var DefaultInclude = []string{
	".bash_aliases",
	".bashrc",
	".dircolors",
	".gitconfig",
	".nano/",
	".nanorc",
	".oh-my-zsh/plugins/",
	".profile",
	".selected_editor",
	".zshrc",
	".zprofile",
	".zlogin",
	".zlogout",
}

func DefaultUsers() map[string]User {
	return map[string]User{
		"root":     {Home: "/root/"},
		"cjserver": {Home: "/home/cjserver/"},
	}
}
