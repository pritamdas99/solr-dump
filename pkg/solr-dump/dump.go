package solr_dump

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pritamdas99/solr-dump/blob"
	"github.com/pritamdas99/solr-dump/model"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	dbc "kubedb.dev/db-client-go/solr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

var (
	scm = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientSetScheme.AddToScheme(scm))
	utilruntime.Must(api.AddToScheme(scm))
}

type SolrDump struct {
	action     string
	slClient   dbc.SLClient
	location   string
	repository string
}

func NewSolrDump(action string, dbname string, namespace string, location string, repository string) (*SolrDump, error) {
	if action != "restore" {
		action = "backup"
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("Failed to get config %s", config)
		return nil, err
	}
	kc, err := client.New(config, client.Options{
		Scheme: scm,
		Mapper: nil,
	})
	if err != nil {
		fmt.Println("failed to get client")
		return nil, err
	}
	db := &api.Solr{}
	err = kc.Get(context.TODO(), types.NamespacedName{
		Name:      dbname,
		Namespace: namespace,
	}, db)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	slClient, err := dbc.NewKubeDBClientBuilder(kc, db).WithContext(ctx).WithLog(klog.Background()).GetSolrClient()
	if err != nil {
		return nil, err
	}
	return &SolrDump{
		action,
		slClient,
		location,
		repository,
	}, nil
}

func (dumper *SolrDump) flushStatus(asyncId string) error {
	resp, err := dumper.slClient.FlushStatus(asyncId)
	if err != nil {
		return err
	}

	responseBody, err := dumper.slClient.DecodeResponse(resp)
	if err != nil {
		return err
	}

	_, err = dumper.slClient.GetResponseStatus(responseBody)
	if err != nil {
		klog.Error(fmt.Sprintf("status is non zero while listing collection"))
		return err
	}

	return nil
}

func (dumper *SolrDump) Execute() {
	if dumper.action == "backup" {
		err := dumper.backup()
		if err != nil {
			klog.Error(err)
		}
	} else {
		err := dumper.restore()
		if err != nil {
			klog.Error(err)
		}
	}
}

func (dumper *SolrDump) checkStatus(collections []string) int {
	fl := 0

	for _, collection := range collections {
		if collection == "kubedb-system" {
			continue
		}
		asyncId := fmt.Sprintf("%s-%s", collection, dumper.action)
		resp, err := dumper.slClient.RequestStatus(asyncId)
		if err != nil {
			klog.Error(fmt.Sprintf("Failed to get response for asyncId %s. Error: %v", asyncId, err))
			continue
		}

		responseBody, err := dumper.slClient.DecodeResponse(resp)
		if err != nil {
			klog.Error(fmt.Sprintf("Failed to decode response for asyncId %s. Error: %v", asyncId, err))
			continue
		}

		_, err = dumper.slClient.GetResponseStatus(responseBody)
		if err != nil {
			klog.Error(fmt.Sprintf("status is non zero while checking status for asyncId %s. Error %v", asyncId, err))
			continue
		}

		state, err := dumper.slClient.GetAsyncStatus(responseBody)
		if err != nil {
			klog.Error(fmt.Sprintf("status is non zero while checking state of async for asyncId %s. Error %v", asyncId, err))
			continue
		}
		if state == "completed" {
			klog.Info(fmt.Sprintf("API call for asyncId %s completed.", asyncId))
			err := dumper.flushStatus(asyncId)
			if err != nil {
				klog.Error(fmt.Sprintf("Failed to flush api call for asyncId %s. Error %v", asyncId, err))
			}
			collection = "kubedb-system"
		} else if state == "failed" {
			klog.Info(fmt.Sprintf("API call for asyncId %s failed", asyncId))
			err := dumper.flushStatus(asyncId)
			if err != nil {
				klog.Error(fmt.Sprintf("Failed to flush api call for asyncId %s. Error %v", asyncId, err))
			}
			collection = "kubedb-system"
		} else if state == "notfound" {
			klog.Info(fmt.Sprintf("API call for asyncid %s not found", asyncId))
			collection = "kubedb-system"
		} else {
			fl = 1
		}
	}
	return fl
}

func (dumper *SolrDump) backup() error {
	resp, err := dumper.slClient.ListCollection()
	if err != nil {
		return err
	}

	responseBody, err := dumper.slClient.DecodeResponse(resp)
	if err != nil {
		return err
	}

	_, err = dumper.slClient.GetResponseStatus(responseBody)
	if err != nil {
		klog.Error(fmt.Sprintf("status is non zero while listing collection"))
		return err
	}

	collectionList, err := dumper.slClient.GetCollectionList(responseBody)
	if err != nil {
		return err
	}

	g := new(errgroup.Group)

	for _, collection := range collectionList {
		if collection == "kubedb-system" {
			continue
		}
		g.Go(func() error {
			klog.Info(fmt.Sprintf("BACKUP COLLECTION %s\n", collection))
			resp, err := dumper.slClient.BackupCollection(context.TODO(), collection, fmt.Sprintf("%s-backup", collection), dumper.location, dumper.repository)
			if err != nil {
				klog.Error(fmt.Sprintf("Failed to backup collection %s", collection))
				return err
			}
			responseBody, err := dumper.slClient.DecodeResponse(resp)
			if err != nil {
				klog.Error(fmt.Sprintf("Failed to decode backup response body for collection %s", collection))
				return err
			}
			_, err = dumper.slClient.GetResponseStatus(responseBody)
			if err != nil {
				klog.Error(fmt.Sprintf("status is non zero while listing collection"))
				return err
			}

			b, err := json.MarshalIndent(responseBody, "", " ")
			if err != nil {
				klog.Error(fmt.Sprintf("Could not format response for collection %s into json", collection))
			}
			klog.Infof(fmt.Sprintf("responsebody %v", string(b)))
			return nil
		})
	}

	if err := g.Wait(); err == nil {
		fmt.Println("Successfully took backup.")
	}

	return nil
}

func NewS3(prefix string) (model.Blob, error) {
	bs := &model.BackupStorage{
		Storage: model.Storage{
			Provider: model.ProviderS3,
			S3: &model.S3{
				Bucket:   "solrbackup",
				Region:   "us-east-1",
				Endpoint: "http://proxy-svc.demo.svc:80",
				Prefix:   prefix,
			},
		},
	}
	return blob.NewBlob(bs)
}

func (dumper *SolrDump) restore() error {

	s3b, err := NewS3("/")
	if err != nil {
		return err
	}

	list, err := s3b.List(context.TODO(), "/")
	if err != nil {
		return err
	}
	backupName := ""
	collection := ""
	var arr [][2]string
	for _, x := range list {
		part := strings.Split(strings.Trim(x, "/"), "/")
		if part[0] == backupName || part[1] == collection {
			continue
		}
		backupName = part[0]
		collection = part[1]
		ar := [2]string{part[0], part[1]}
		arr = append(arr, ar)
	}
	for _, x := range arr {
		fmt.Println(x)
	}
	g := new(errgroup.Group)
	for _, x := range arr {
		g.Go(func() error {
			backupName := x[0]
			collection := x[1]
			klog.Info(fmt.Sprintf("RESTORE COLLECTION %s\n", collection))
			klog.Info(fmt.Sprintf("an api call with %s %s gone", collection, backupName))
			resp, err := dumper.slClient.RestoreCollection(context.TODO(), collection, backupName, dumper.location, dumper.repository)
			if err != nil {
				klog.Error(fmt.Sprintf("Failed to backup collection %s", collection))
				return err
			}
			responseBody, err := dumper.slClient.DecodeResponse(resp)
			if err != nil {
				klog.Error(fmt.Sprintf("Failed to decode backup response body for collection %s", collection))
				return err
			}
			_, err = dumper.slClient.GetResponseStatus(responseBody)
			if err != nil {
				klog.Error(fmt.Sprintf("status is non zero while restore collection %s\n", collection))
				return err
			}
			b, err := json.MarshalIndent(responseBody, "", " ")
			if err != nil {
				klog.Error(fmt.Sprintf("Could not format response for collection %s into json", collection))
			}
			klog.Infof(fmt.Sprintf("responsebody %v", string(b)))
			return nil
		})
	}

	if err := g.Wait(); err == nil {
		fmt.Println("Successfully restored all collections.")
	}

	return nil
}
