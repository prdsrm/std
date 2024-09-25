package dialogs

import (
	"cmp"
	"context"
	"slices"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
	"github.com/prdsrm/std/utils"
)

func GetDialogs(ctx context.Context, client *telegram.Client) ([]dialogs.Elem, error) {
	builder := query.GetDialogs(client.API())
	dialogs, err := builder.Collect(ctx)
	if err != nil {
		return nil, err
	}
	return dialogs, nil
}

func SortPeersInDialogs(elems []dialogs.Elem) []tg.PeerClass {
	var peers []tg.PeerClass
	for _, elem := range elems {
		peers = append(peers, elem.Dialog.GetPeer())
	}
	slices.SortFunc(peers, func(firstPeerClass tg.PeerClass, secondPeerClass tg.PeerClass) int {
		id1 := utils.GetIDFromPeerClass(firstPeerClass)
		id2 := utils.GetIDFromPeerClass(secondPeerClass)
		return cmp.Compare(id1, id2)
	})
	return peers
}

func compare(peerClass tg.PeerClass, id int64) int {
	return cmp.Compare(utils.GetIDFromPeerClass(peerClass), id)
}

func SearchTroughPeers(peers []tg.PeerClass, id int64) (*tg.PeerClass, bool) {
	index, found := slices.BinarySearchFunc(peers, id, compare)
	if found {
		return &peers[index], true
	}
	return nil, false
}

// removed the wrapper
// taken from https://github.com/celestix/gotgproto
func ArchiveChats(
	ctx context.Context,
	client *telegram.Client,
	peers []tg.InputPeerClass,
) (bool, error) {
	var folderPeers = make([]tg.InputFolderPeer, len(peers))
	for n, peer := range peers {
		folderPeers[n] = tg.InputFolderPeer{
			Peer:     peer,
			FolderID: 1,
		}
	}
	_, err := client.API().FoldersEditPeerFolders(ctx, folderPeers)
	return err == nil, err
}
