package akeso

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/googlecloudplatform/gcsfuse/v2/internal/logger"
)

// TODO: dynamic strategy based on passed config
const Strategy = "akeso"

type Config struct {
	Strategy  string
	AkesoDir  string
	ProjectID string
	// TopicID   string 
	SetupGroupTopicID	string
	UpdateKeyTopicID	string
	// SubID     string 
	// Subscription IDs (= topic_name + member_name + "sub") to be generated and created automatically
	MemberName string
	Key       []byte
	KeyMutex  sync.RWMutex

	ArtConfig ArtConfig
}

func (c *Config) String() string {
	return fmt.Sprintf("AkesoConfig{Strategy: %q, AkesoDir: %q, ProjectId: %q, MemberName: %q, SetupGroupTopicID: %q, UpdateKeyTopicID: %q}",
		c.Strategy, c.AkesoDir, c.ProjectID, c.MemberName, c.SetupGroupTopicID, c.UpdateKeyTopicID)
}

func (c *Config) SetKey(key []byte) {
	c.KeyMutex.Lock()
	defer c.KeyMutex.Unlock()
	c.Key = key
}

type ArtConfig struct {
	index      int
	memberName string // member name is hash of member's public key/email/name
	groupName  string // group name is a group id or a group name

	// keys and setup files are only created once during the setup phase
	initiatorPubIKFile string
	memberPrivEKFile   string

	setupMsgFile    string
	setupMsgSigFile string

	updateMsgFile    string
	updateMsgMacFile string

	// treeStateFile and stageKeyFile are updated by all the setup and update messages
	treeStateFile string
	// there will be multiple stage key files
	// rename the current stageKeyFile to something else
	// then write the stageKeyFile
	stageKeyFile string

	// serialize pubsub messages for debugging
	pubsubMsgDir string

	keyUpdateTopicID string
	keyUpdateSubID   string
}

func DefaultArtConfig(akesoDir string) *ArtConfig {
	artConfig := &ArtConfig{}

	// how does group member know their index in the group?
	// artConfig.index = 3
	// artConfig.memberName = "cici"
	artConfig.index = 2
	artConfig.memberName = "bob"
	artConfig.groupName = "abcd"

	// artConfig.keyUpdateSubID = "KeyUpdate-cici"
	artConfig.keyUpdateSubID = "KeyUpdate-bob"
	artConfig.keyUpdateTopicID = "KeyUpdate"

	baseDir := filepath.Join(akesoDir, artConfig.groupName, artConfig.memberName)
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
	artConfig.initiatorPubIKFile = filepath.Join(keysDir, "initiator-ik-pub.pem")
	artConfig.memberPrivEKFile = filepath.Join(keysDir, "member-ek.pem")

	// smh/akeso.d/{groupName}/{memberName}/setup/
	// smh/akeso.d/{groupName}/{memberName}/setup/setup.msg // setupMsgFile
	// smh/akeso.d/{groupName}/{memberName}/setup/setup.msg.sig // setupMsgSigFile
	artConfig.setupMsgFile = filepath.Join(setupDir, "setup.msg")
	artConfig.setupMsgSigFile = filepath.Join(setupDir, "setup.msg.sig")

	// smh/akeso.d/{groupName}/{memberName}/state.json // treeStateFile
	// smh/akeso.d/{groupName}/{memberName}/stage-key.pem // stageKeyFile
	artConfig.treeStateFile = filepath.Join(baseDir, "state.json")
	artConfig.stageKeyFile = filepath.Join(baseDir, "stage-key.pem")

	// smh/akeso.d/{groupName}/{memberName}/pubsub // pubsubMsgDir
	// smh/akeso.d/{groupName}/{memberName}/pubsub/{publishTime}-{msgID}.json
	artConfig.pubsubMsgDir = pubsubMsgDir

	// smh/akeso.d/{groupName}/{memberName}/update/
	// smh/akeso.d/{groupName}/{memberName}/update/update.msg
	// smh/akeso.d/{groupName}/{memberName}/update/update.msg.mac
	artConfig.updateMsgFile = filepath.Join(updateDir, "update.msg")
	artConfig.updateMsgMacFile = filepath.Join(updateDir, "update.msg.mac")

	return artConfig
}
