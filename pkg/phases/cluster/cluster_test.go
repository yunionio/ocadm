package cluster

import "testing"

func Test_getRepoImageName(t *testing.T) {
	type args struct {
		img string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{
			"yunion/image",
			args{"yunion/image"},
			"yunion",
			"image",
			"latest",
		},
		{
			"hub.docker.io/yunion/image:v2",
			args{"hub.docker.io/yunion/image:v2"},
			"hub.docker.io/yunion",
			"image",
			"v2",
		},
		{
			"hub.docker.io/yunion/image:v2",
			args{"hub.docker.io/yunion/image:v2"},
			"hub.docker.io/yunion",
			"image",
			"v2",
		},
		{
			"10.168.222.173:8082/yunion/image:v2",
			args{"10.168.222.173:8082/yunion/image:v2"},
			"10.168.222.173:8082/yunion",
			"image",
			"v2",
		},
		{
			"registry.cn-beijing.aliyuncs.com/yunionio/climc@sha256:32dcddaa6271b8c752bd6574c789771d38315076ce300c4a1d4618496e359f2d",
			args{"registry.cn-beijing.aliyuncs.com/yunionio/climc@sha256:32dcddaa6271b8c752bd6574c789771d38315076ce300c4a1d4618496e359f2d"},
			"registry.cn-beijing.aliyuncs.com/yunionio",
			"climc",
			"sha256:32dcddaa6271b8c752bd6574c789771d38315076ce300c4a1d4618496e359f2d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, _ := getRepoImageName(tt.args.img)
			if got != tt.want {
				t.Errorf("getRepoImageName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getRepoImageName() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("getRepoImageName() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
