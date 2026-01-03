package config

var DefaultTrack = []string{
	".agent",
	".bash_aliases",
	".bashrc",
	".colors",
	".config/glow/glow.yml",
	".dircolors",
	".fzf_git",
	".gitconfig",
	".nano/**",
	".nanorc",
	".oh-my-zsh/plugins/**/*.zsh",
	".oh-my-zsh/themes/*.zsh-theme",
	".profile",
	".selected_editor",
	".wakeonlan",
	".zlogin",
	".zlogout",
	".zprofile",
	".zshrc",
	"docs/",
	"!docs/config.yaml",
}

func DefaultUsers() map[string]User {
	return map[string]User{
		"root":   {Home: "/root/"},
		"ubuntu": {Home: "/home/ubuntu/"},
	}
}
