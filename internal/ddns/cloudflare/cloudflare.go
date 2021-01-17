package cloudflare

import (
	"fmt"
	"net"
	"os"

	"k8s.io/klog/v2"

	"github.com/USA-RedDragon/redwall/internal/iplistener"
	"github.com/cloudflare/cloudflare-go"
)

type CloudflareDDNS struct {
	// required
	currentIP  net.IP
	ipv6       bool
	iplistener *iplistener.IPListener
	// do not set
	cfAPI      *cloudflare.API
	zoneID     string
	cfAPIToken string
	cfRecord   string
	recordType string
}

func (c *CloudflareDDNS) getCurrentRecordValue() (net.IP, string) {
	currentRecords, err := c.cfAPI.DNSRecords(
		c.zoneID,
		cloudflare.DNSRecord{Name: c.cfRecord, Type: c.recordType},
	)
	if err != nil {
		klog.Error(err)
	}

	if len(currentRecords) > 1 {
		klog.Errorf("Multiple type %s records returned.", c.recordType)
		return net.ParseIP(currentRecords[0].Content), currentRecords[0].ID
	} else if len(currentRecords) == 0 {
		klog.Info("No existing record found.")
		return nil, ""
	}

	return net.ParseIP(currentRecords[0].Content), currentRecords[0].ID
}

func (c *CloudflareDDNS) deleteRecord() {
	_, currentRecordID := c.getCurrentRecordValue()
	if currentRecordID != "" {
		err := c.cfAPI.DeleteDNSRecord(c.zoneID, currentRecordID)
		if err != nil {
			klog.Error(err)
		} else {
			klog.Info("Removed DDNS entry")
		}
	}
}

func (c *CloudflareDDNS) updateDNS(newIP net.IP) {
	klog.Infof("Received IP update: %v", newIP)

	if newIP == nil {
		c.deleteRecord()
		return
	}

	needCreateRecord := false
	currentRecord, currentRecordID := c.getCurrentRecordValue()
	if currentRecord == nil {
		needCreateRecord = true
	}

	if needCreateRecord || !newIP.Equal(currentRecord) {
		klog.Info("Updating record due to IP mismatch")

		newRecord := cloudflare.DNSRecord{
			Name:    c.cfRecord,
			Content: newIP.String(),
			Type:    c.recordType,
			TTL:     120,
		}
		if !needCreateRecord {
			newRecord.ID = currentRecordID
		}

		err := c.upsertIP(&newRecord, needCreateRecord)
		if err != nil {
			klog.Error(err)
		} else {
			klog.Info("Updated DNS record")
		}
	}
}

func (c *CloudflareDDNS) upsertIP(newRecord *cloudflare.DNSRecord, create bool) error {
	if create {
		rr, err := c.cfAPI.CreateDNSRecord(c.zoneID, *newRecord)
		if err != nil {
			return err
		}
		if !rr.Response.Success {
			klog.Error("Failed to create record: %v", rr.Response)
		}
	} else {
		err := c.cfAPI.UpdateDNSRecord(c.zoneID, newRecord.ID, *newRecord)
		if err != nil {
			return err
		}
	}
	return nil
}

// Start the DDNS service
func (c *CloudflareDDNS) Start() {
	c.updateDNS(c.currentIP)
	c.iplistener.Subscribe(c.updateDNS)
}

func New(currentIP net.IP, ipv6 bool, iplistener *iplistener.IPListener) *CloudflareDDNS {
	cfAPIToken := os.Getenv("CF_API_TOKEN")
	cfAPI, err := cloudflare.NewWithAPIToken(cfAPIToken)
	if err != nil {
		klog.Error(err)
		return nil
	}

	cfZone := os.Getenv("CF_ZONE")
	zoneID, err := cfAPI.ZoneIDByName(cfZone)
	if err != nil {
		klog.Error(err)
		return nil
	}

	envCFRecord := os.Getenv("CF_RECORD")
	if envCFRecord == "" {
		klog.Error("Must provide CF_RECORD with the subdomain to set with DDNS")
		return nil
	}
	cfRecord := fmt.Sprintf("%s.%s", envCFRecord, cfZone)

	var recordType string
	if ipv6 {
		recordType = "AAAA"
	} else {
		recordType = "A"
	}

	return &CloudflareDDNS{
		currentIP:  currentIP,
		ipv6:       ipv6,
		iplistener: iplistener,
		cfAPI:      cfAPI,
		zoneID:     zoneID,
		cfAPIToken: cfAPIToken,
		recordType: recordType,
		cfRecord:   cfRecord,
	}
}
