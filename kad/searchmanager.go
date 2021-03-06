package kad

import (
	"hahajing/com"
	"log"
	"time"
)

// SearchManager x
type SearchManager struct {
	pPacketProcessor *PacketProcessor

	searchCount uint64
	searchMap   map[[16]byte][]*Search // key is 128bits KAD hash of keyword

	decision SearchDecision
}

func (sm *SearchManager) start(pPacketProcessor *PacketProcessor, pOnliner *ContactOnliner) {
	sm.pPacketProcessor = pPacketProcessor

	sm.searchMap = make(map[[16]byte][]*Search)

	sm.decision.start(pOnliner)
}

func (sm *SearchManager) goSearch(pSearch *Search) {
	com.HhjLog.Infof("New search: %s", pSearch.keyword)

	contacts := sm.decision.newSearch(pSearch)
	pSearch.goSearch(contacts, sm.pPacketProcessor)
}

func (sm *SearchManager) newSearch(pSearchReq *SearchReq) {
	for _, keyword := range pSearchReq.Keywords {
		no := sm.searchCount
		sm.searchCount++

		targetHash := sm.getKeywordHash(keyword)

		search := Search{
			no:           no,
			respCh:       pSearchReq.RespCh,
			Keywords:     pSearchReq.Keywords,
			targetID:     ID{hash: targetHash},
			keyword:      keyword,
			tExpires:     time.Now().Unix() + searchExpires,
			fileHashMap:  make(map[[16]byte]bool),
			contactIPMap: make(map[uint32]bool)}

		searches := append(sm.searchMap[targetHash], &search)
		sm.searchMap[targetHash] = searches

		if len(searches) == 1 { // we're the first one
			sm.goSearch(&search)
		} else {
			log.Printf("Ongoing search: %s", keyword)

			// There's same target search ongoing, get the first one
			pSearch := searches[0]

			// send matched files to user for each search
			fileLinks := pSearch.convert2FileLinks(pSearch.files)
			if fileLinks != nil {
				if len(pSearch.respCh) < cap(pSearch.respCh) {
					searchRes := SearchResp{FileLinks: fileLinks}
					pSearch.respCh <- &searchRes
				}
			}
		}
	}
}

func (sm *SearchManager) getKeywordHash(keyword string) [16]byte {
	md4 := Md4Sum{}
	md4.calculate([]byte(keyword))

	// change to big endian for each uint32
	hash := com.ConvertEd2kHash32(md4.getRawHash())

	return hash
}

func (sm *SearchManager) addKademlia2SearchResp(pMsg *Kademlia2SearchRespMsg) {
	targetHash := pMsg.targetID.get()
	searches := sm.searchMap[targetHash]
	if searches == nil {
		return
	}

	// add files into the first one
	pSearch := searches[0]
	newFiles := pSearch.addFiles(pMsg.files)
	if newFiles == nil {
		return
	}

	// send new file links to user for each search
	for _, pSearch := range searches {
		// convert to file links matched with user search keywords
		fileLinks := pSearch.convert2FileLinks(newFiles)

		if fileLinks != nil {
			if len(pSearch.respCh) < cap(pSearch.respCh) {
				searchRes := SearchResp{FileLinks: fileLinks}
				pSearch.respCh <- &searchRes
			}
		}
	}
}

func (sm *SearchManager) addKademlia2Resp(pMsg *Kademlia2RespMsg) bool {
	// check if this is our target
	hash := pMsg.targetID.get()
	searches := sm.searchMap[hash]
	if searches == nil {
		return false
	}

	// continue search
	pSearch := searches[0]
	pSearch.goSearch(pMsg.contacts, sm.pPacketProcessor)

	return true
}

func (sm *SearchManager) tickProcess() {
	t := time.Now().Unix()

	for key, searches := range sm.searchMap {
		pSearch := searches[len(searches)-1]
		if t >= pSearch.tExpires {
			delete(sm.searchMap, key)
		}
	}
}
