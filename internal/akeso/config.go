package akeso

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/googlecloudplatform/gcsfuse/v2/internal/config"
	"github.com/googlecloudplatform/gcsfuse/v2/internal/logger"
)

// TODO: dynamic strategy based on passed config
const Strategy = "akeso"

type Config struct {
	config.AkesoConfig

	Key      []byte
	KeyMutex sync.RWMutex

	PubSubDir string
}

func (c *Config) String() string {
	return fmt.Sprintf("AkesoConfig{Strategy: %q, AkesoDir: %q, ProjectId: %q, MemberName: %q, SetupGroupTopicID: %q, UpdateKeyTopicID: %q}",
		c.Strategy, c.AkesoDir, c.ProjectID, c.ArtConfig.MemberName, c.SetupTopicID, c.UpdateTopicID)
}

func (c *Config) SetKeyFile(stageKeyFile string) {
	salt := []byte("YOUR_RANDOM_SALT")
	key, err := AESFromPEM(stageKeyFile, salt)
	if err != nil {
		logger.Errorf("error: %v", err)
	}

	keyFile := filepath.Join(c.AkesoDir, "key")
	err = os.WriteFile(keyFile, key, 0666)
	if err != nil {
		logger.Errorf("error writing key file: %v", err)
	}

	c.SetKey(key)
}

func (c *Config) SetKey(key []byte) {
	c.KeyMutex.Lock()
	defer c.KeyMutex.Unlock()
	c.Key = key
}

func NewAkesoConfig(mountConfig *config.MountConfig) *Config {
	akesoConfig := &Config{}
	akesoConfig.Strategy = mountConfig.Strategy
	akesoConfig.AkesoDir = mountConfig.AkesoDir
	akesoConfig.ProjectID = mountConfig.ProjectID

	akesoConfig.SetupTopicID = mountConfig.SetupTopicID
	akesoConfig.SetupSubID = mountConfig.SetupSubID
	akesoConfig.UpdateTopicID = mountConfig.UpdateTopicID
	akesoConfig.UpdateSubID = mountConfig.UpdateSubID

	akesoConfig.KeyFile = mountConfig.KeyFile

	artConfig := NewArtConfig(mountConfig)
	akesoConfig.ArtConfig = *artConfig
	akesoConfig.PubSubDir = artConfig.PubSubDir

	return akesoConfig
}

func NewArtConfig(mountConfig *config.MountConfig) *config.ArtConfig {
	artConfig := &config.ArtConfig{}

	akesoConfig := &mountConfig.AkesoConfig
	ac := mountConfig.AkesoConfig.ArtConfig

	// how does group member know their index in the group?
	// artConfig.index = 3
	// artConfig.memberName = "cici"
	// artConfig.memberName = "abcd"
	artConfig.Index = ac.Index
	artConfig.MemberName = ac.MemberName
	artConfig.GroupName = ac.GroupName

	baseDir := filepath.Join(akesoConfig.AkesoDir, ac.GroupName, ac.MemberName)
	keysDir := filepath.Join(baseDir, "keys")
	setupDir := filepath.Join(baseDir, "setup")
	pubsubMsgDir := filepath.Join(baseDir, "pubsub")
	updateDir := filepath.Join(baseDir, "update") // stores update messages

	err := CreateDirsIfNotExist(baseDir, keysDir, setupDir, updateDir, pubsubMsgDir)
	if err != nil {
		logger.Errorf("util.CreateDirsIfNotExist failed: %v", err)
	}

	// smh/akeso.d/{groupName}/{memberName}/keys/
	// smh/akeso.d/{groupName}/{memberName}/keys/initiator-ik-pub.pem // initiatorPubIKFile
	// smh/akeso.d/{groupName}/{memberName}/keys/member-ek.pem // memberPrivEKFile
	artConfig.InitiatorPubIKFile = filepath.Join(keysDir, ac.InitiatorPubIKFile)
	artConfig.MemberPrivEKFile = filepath.Join(keysDir, ac.MemberPrivEKFile)

	// smh/akeso.d/{groupName}/{memberName}/setup/
	// smh/akeso.d/{groupName}/{memberName}/setup/setup.msg // setupMsgFile
	// smh/akeso.d/{groupName}/{memberName}/setup/setup.msg.sig // setupMsgSigFile
	artConfig.SetupMsgFile = filepath.Join(setupDir, ac.SetupMsgFile)
	artConfig.SetupMsgSigFile = filepath.Join(setupDir, ac.SetupMsgSigFile)

	// smh/akeso.d/{groupName}/{memberName}/state.json // treeStateFile
	// smh/akeso.d/{groupName}/{memberName}/stage-key.pem // stageKeyFile
	artConfig.TreeStateFile = filepath.Join(baseDir, ac.TreeStateFile)
	artConfig.StageKeyFile = filepath.Join(baseDir, ac.StageKeyFile)

	// smh/akeso.d/{groupName}/{memberName}/pubsub // pubsubMsgDir
	// smh/akeso.d/{groupName}/{memberName}/pubsub/{publishTime}-{msgID}.json
	artConfig.PubSubDir = pubsubMsgDir

	// smh/akeso.d/{groupName}/{memberName}/update/
	// smh/akeso.d/{groupName}/{memberName}/update/update.msg
	// smh/akeso.d/{groupName}/{memberName}/update/update.msg.mac
	artConfig.UpdateMsgFile = filepath.Join(updateDir, ac.UpdateMsgFile)
	artConfig.UpdateMsgMacFile = filepath.Join(updateDir, ac.UpdateMsgMacFile)

	return artConfig
}
