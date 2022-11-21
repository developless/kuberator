package util

import (
	"crypto/md5"
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	corev1 "k8s.io/api/core/v1"
	"sort"
	"strings"
)

func BuildConfResource(cfs []*v1.NamedProperties, components []*v1.CategoryClusterComponent) []*core.CategoryComponentObject {
	var cms []*core.CategoryComponentObject
	if components != nil && len(components) > 0 {
		for _, com := range components {
			var conf = MergeConf(cfs, com.Properties)
			if conf == nil || len(conf) == 0 {
				continue
			}
			cm := &core.CategoryComponentObject{
				CommonCategoryComponent: v1.CommonCategoryComponent{
					Category: ConfigMap,
					Name:     v1.ComponentName(GetConfigMapName(com.GetCategory())),
					Component: v1.Component{
						Kind: ConfigMap,
					},
				},
				Object:    conf,
				Reference: com,
			}
			cms = append(cms, cm)
		}
	}
	return cms
}

// TODO: 根据conf的path拆分成多个configmap，每个configmap mount到conf的path上.
func BuildConfigMapMount(category v1.Category, clusterName string, properties []v1.NamedProperties) ([]corev1.Volume, []corev1.VolumeMount) {
	if properties == nil || len(properties) == 0 {
		return nil, nil
	}
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	var confVol = corev1.Volume{
		Name: fmt.Sprintf("%s-%s", AppConfigMapVolume, category),
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: GetComponentName(clusterName, category, ConfigMap),
				},
			},
		},
	}

	var confMount = corev1.VolumeMount{
		Name:      fmt.Sprintf("%s-%s", AppConfigMapVolume, category),
		MountPath: properties[0].Path,
	}

	for i := range properties {
		confVol.ConfigMap.Items = append(confVol.ConfigMap.Items, corev1.KeyToPath{
			Key:  properties[i].Name,
			Path: properties[i].Name,
		})
	}

	volumes = append(volumes, confVol)
	volumeMounts = append(volumeMounts, confMount)
	return volumes, volumeMounts
}

// MergeConf merge from base.
func MergeConf(base []*v1.NamedProperties, component []*v1.NamedProperties) []*v1.NamedProperties {
	// Properties which should be provided from real deployed environment.
	conf := map[string]*v1.NamedProperties{}
	if base != nil {
		for _, c := range base {
			conf[c.PropertiesName()] = c
		}
	}

	// component conf
	if component != nil {
		if conf == nil {
			conf = map[string]*v1.NamedProperties{}
		}
		for _, v := range component {
			conf[v.PropertiesName()] = v
			//if conf[v.PropertiesName()] == nil {
			//	conf[v.PropertiesName()] = v
			//} else {
			//	p := map[string]string{}
			//	b := conf[v.PropertiesName()]
			//	for bk, bv := range b.Data {
			//		p[bk] = bv
			//	}
			//	for ck, cv := range v.Data {
			//		p[ck] = cv
			//	}
			//	conf[v.PropertiesName()].Data = p
			//}
		}
	}

	var confList []*v1.NamedProperties
	for _, v := range conf {
		confList = append(confList, v)
	}

	return confList
}

func GetConfigMapName(category v1.Category) string {
	return strings.ToLower(fmt.Sprintf("%s-%s", category, ConfigMap))
}

func ConfFinger(confList []v1.NamedProperties) string {
	keys := make([]string, len(confList))
	i := 0
	fingerMap := map[string]string{}
	for _, c := range confList {
		keys[i] = c.PropertiesName()
		fingerMap[keys[i]] = fmt.Sprintf("%x", md5.Sum([]byte(c.Data)))
		i = i + 1
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(fingerMap[key])
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(builder.String())))
}
