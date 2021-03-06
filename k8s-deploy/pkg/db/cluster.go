/*
* Copyright 2019-2020 VMware, Inc.
* 
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
* http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
* 
*/
package db

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type Cluster struct {
	Uuid      string `json:"uuid"`
	Name      string `json:"name"`
	NameSpace string `json:"namespaces"`
	// Cluster revision
	Revision int `json:"revision"`
	// Helm chart version, example: fate v1.2.0
	ChartVersion string                 `json:"chart_version"`
	ChartValues  map[string]interface{} `json:"chart_values"`
	// The value of this cluster for installing helm chart
	Values    string                 `json:"values"`
	ChartName string                 `json:"chart_name"`
	Type      string                 `json:"cluster_type,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	Status    ClusterStatus          `json:"status"`
	//Backend          ComputingBackend       `json:"backend"`
	//BootstrapParties Party                  `json:"bootstrap_parties"`
	Config map[string]interface{} `json:"Config,omitempty"`
	Info   map[string]interface{} `json:"Info,omitempty"`
}

type ClusterStatus int

const (
	Creating_c ClusterStatus = iota
	Deleting_c
	Updating_c
	Running_c
	Unavailable_c
	Deleted_c
)

func (s ClusterStatus) String() string {
	names := []string{
		"Creating",
		"Deleting",
		"Updating",
		"Running",
		"Unavailable",
		"Deleted",
	}

	return names[s]
}

// MarshalJSON convert cluster status to string
func (s *ClusterStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON sets *m to a copy of data.
func (s *ClusterStatus) UnmarshalJSON(data []byte) error {
	if s == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	var ClusterStatus ClusterStatus
	switch string(data) {
	case "\"Creating\"":
		ClusterStatus = Creating_c
	case "\"Deleting\"":
		ClusterStatus = Deleting_c
	case "\"Updating\"":
		ClusterStatus = Updating_c
	case "\"Running\"":
		ClusterStatus = Running_c
	case "\"Unavailable\"":
		ClusterStatus = Unavailable_c
	case "\"Deleted\"":
		ClusterStatus = Deleted_c
	default:
		return errors.New("data can't UnmarshalJSON")
	}

	//log.Debug().Interface("JobStatus", JobStatus).Bytes("datab", data).Str("data", string(data)).Msg("UnmarshalJSON")
	*s = ClusterStatus
	return nil
}

func (cluster *Cluster) getCollection() string {
	return "cluster"
}

// GetUuid get cluster uuid
func (cluster *Cluster) GetUuid() string {
	return cluster.Uuid
}

// FromBson convert bson to cluster
func (cluster *Cluster) FromBson(m *bson.M) (interface{}, error) {
	bsonBytes, err := bson.Marshal(m)
	if err != nil {
		return nil, err
	}
	err = bson.Unmarshal(bsonBytes, cluster)
	if err != nil {
		return nil, err
	}
	return *cluster, nil
}

// NewCluster create cluster object with basic argument
func NewCluster(name string, nameSpaces, chartName, chartVersion string) *Cluster {
	cluster := &Cluster{
		Uuid:         uuid.NewV4().String(),
		Name:         name,
		NameSpace:    nameSpaces,
		Revision:     0,
		Status:       Creating_c,
		ChartName:    chartName,
		ChartVersion: chartVersion,
	}

	return cluster
}

// ClusterFindByUUID get cluster from via uuid
func ClusterFindByUUID(uuid string) (*Cluster, error) {
	result, err := FindOneByUUID(new(Cluster), uuid)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("cluster no find")
	}
	Cluster, ok := result.(Cluster)
	if !ok {
		return nil, errors.New("assertion type error")
	}
	log.Debug().Interface("Cluster", Cluster).Msg("find Cluster success")
	return &Cluster, nil
}

// ClusterFindByName get cluster from via name
func ClusterFindByName(name, namespace string) (*Cluster, error) {

	filter := bson.M{"name": name, "namespace": namespace, "status": Running_c}
	result, err := FindOneByFilter(new(Cluster), filter)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("cluster no find")
	}
	Cluster, ok := result.(Cluster)
	if !ok {
		return nil, errors.New("assertion type error")
	}
	log.Debug().Interface("Cluster", Cluster).Msg("find Cluster success")
	return &Cluster, nil
}

// FindClusterList get all cluster list
func FindClusterList(args string, all bool) ([]*Cluster, error) {

	cluster := &Cluster{}
	filter := bson.M{}
	result, err := FindByFilter(cluster, filter)
	if err != nil {
		return nil, err
	}

	clusterList := make([]*Cluster, 0)
	for _, r := range result {
		cluster := r.(Cluster)
		clusterList = append(clusterList, &cluster)
	}
	return clusterList, nil
}

func ClusterDeleteByUUID(uuid string) error {

	cluster, err := ClusterFindByUUID(uuid)
	if err != nil {
		return err
	}
	cluster.Status = Deleted_c
	err = UpdateByUUID(cluster, uuid)
	if err != nil {
		return err
	}

	log.Debug().Interface("ClusterUuid", uuid).Msg("delete Cluster success")
	return nil
}

func (cluster *Cluster) IsExisted(name, namespace string) bool {
	//filter := bson.M{"name": name, "namespace": namespace, "status": bson.M{"$ne": "Deleted"}}
	filter := bson.M{"$and": []bson.M{{"name": name, "namespace": namespace}, {"status": bson.M{"$ne": Deleted_c}}}}
	Clusters, err := FindByFilter(cluster, filter)
	fmt.Println(ToJson(Clusters))
	fmt.Println(err)
	if err != nil || len(Clusters) == 0 {
		return false
	}
	return true
}
