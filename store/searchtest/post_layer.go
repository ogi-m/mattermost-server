package searchtest

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/stretchr/testify/require"
)

var searchPostStoreTests = []searchTest{
	{
		"Should be able to search posts including results from DMs",
		testSearchPostsIncludingDMs,
		[]string{ENGINE_ALL},
	},
	{
		"Should return pinned and unpinned posts",
		testSearchReturnPinnedAndUnpinned,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search for exact phrases in quotes",
		testSearchExactPhraseInQuotes,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search for email addresses with or without quotes",
		testSearchEmailAddresses,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search when markdown underscores are applied",
		testSearchMarkdownUnderscores,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search for non-latin words",
		testSearchNonLatinWords,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search for alternative spellings of words",
		testSearchAlternativeSpellings,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search for alternative spellings of words with and without accents",
		testSearchAlternativeSpellingsAccents,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search or exclude messages written by a specific user",
		testSearchOrExcludePostsBySpecificUser,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search or exclude messages written in a specific channel",
		testSearchOrExcludePostsInChannel,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search or exclude messages written in a DM or GM",
		testSearchOrExcludePostsInDMGM,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to filter messages written on a specific date",
		testFilterMessagesInSpecificDate,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to filter messages written before a specific date",
		testFilterMessagesBeforeSpecificDate,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to filter messages written after a specific date",
		testFilterMessagesAfterSpecificDate,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to exclude messages that contain a serch term",
		testFilterMessagesWithATerm,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search using boolean operators",
		testSearchUsingBooleanOperators,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search with combined filters",
		testSearchUsingCombinedFilters,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to ignore stop words",
		testSearchIgnoringStopWords,
		[]string{ENGINE_ALL},
	},
	{
		"Should support search stemming",
		testSupportStemming,
		[]string{ENGINE_ALL},
	},
	{
		"Should support search with wildcards",
		testSupportWildcards,
		[]string{ENGINE_ALL},
	},
	{
		"Should not support search with preceding wildcards",
		testNotSupportPrecedingWildcards,
		[]string{ENGINE_ALL},
	},
	{
		"Should discard a wildcard if it's not placed immediately by text",
		testSearchDiscardWildcardAlone,
		[]string{ENGINE_ALL},
	},
	{
		"Should support terms with dash",
		testSupportTermsWithDash,
		[]string{ENGINE_ALL},
	},
	{
		"Should support terms with underscore",
		testSupportTermsWithUnderscore,
		[]string{ENGINE_ALL},
	},
	{
		"Should search or exclude post using hashtags",
		testSearchOrExcludePostsWithHashtags,
		[]string{ENGINE_ALL},
	},
	{
		"Should support searching for hashtags surrounded by markdown",
		testSearchHashtagWithMarkdown,
		[]string{ENGINE_ALL},
	},
	{
		"Should support searching for multiple hashtags",
		testSearcWithMultipleHashtags,
		[]string{ENGINE_ALL},
	},
	{
		"Should support searching hashtags with dots",
		testSearchPostsWithDotsInHashtags,
		[]string{ENGINE_ALL},
	},
	{
		"Should be able to search or exclude messages with hashtags in a case insensitive manner",
		testSearchHashtagCaseInsensitive,
		[]string{ENGINE_ALL},
	},
}

func TestSearchPostStore(t *testing.T, s store.Store, testEngine *SearchTestEngine) {
	th := &SearchTestHelper{
		Store: s,
	}
	err := th.SetupBasicFixtures()
	require.Nil(t, err)
	defer th.CleanFixtures()

	runTestSearch(t, testEngine, searchPostStoreTests, th)
}

func testSearchPostsIncludingDMs(t *testing.T, th *SearchTestHelper) {
	direct, err := th.createDirectChannel(th.Team.Id, "direct", "direct", []*model.User{th.User, th.User2})
	require.Nil(t, err)
	defer th.deleteChannel(direct)

	p1, err := th.createPost(th.User.Id, direct.Id, "dm test", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, direct.Id, "dm other", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "channel test", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{Terms: "test"}
	results, err := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, err)

	require.Len(t, results.Posts, 2)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
	th.checkPostInSearchResults(t, p2.Id, results.Posts)
}

func testSearchReturnPinnedAndUnpinned(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "channel test unpinned", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "channel test pinned", "", 0, true)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{Terms: "test"}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 2)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
	th.checkPostInSearchResults(t, p2.Id, results.Posts)
}

func testSearchExactPhraseInQuotes(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "channel test 1 2 3", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "channel test 123", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{Terms: "\"channel test 1 2 3\""}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 1)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
}

func testSearchEmailAddresses(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "test email test@test.com", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "test email test2@test.com", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should search email addresses enclosed by quotes", func(t *testing.T) {
		params := &model.SearchParams{Terms: "\"test@test.com\""}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should search email addresses without quotes", func(t *testing.T) {
		params := &model.SearchParams{Terms: "test@test.com"}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})
}

func testSearchMarkdownUnderscores(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "_start middle end_ _both_", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should search the start inside the markdown underscore", func(t *testing.T) {
		params := &model.SearchParams{Terms: "start"}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should search a word in the middle of the markdown underscore", func(t *testing.T) {
		params := &model.SearchParams{Terms: "middle"}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should search in the end of the markdown underscore", func(t *testing.T) {
		params := &model.SearchParams{Terms: "end"}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should search inside markdown underscore", func(t *testing.T) {
		params := &model.SearchParams{Terms: "both"}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})
}

func testSearchNonLatinWords(t *testing.T, th *SearchTestHelper) {
	t.Run("Should be able to search chinese words", func(t *testing.T) {
		p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "你好", "", 0, false)
		require.Nil(t, err)
		p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "你", "", 0, false)
		require.Nil(t, err)
		defer th.deleteUserPosts(th.User.Id)

		t.Run("Should search one word", func(t *testing.T) {
			params := &model.SearchParams{Terms: "你"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 1)
			th.checkPostInSearchResults(t, p2.Id, results.Posts)
		})
		t.Run("Should search two words", func(t *testing.T) {
			params := &model.SearchParams{Terms: "你好"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 1)
			th.checkPostInSearchResults(t, p1.Id, results.Posts)
		})
		t.Run("Should search with wildcard", func(t *testing.T) {
			params := &model.SearchParams{Terms: "你*"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 2)
			th.checkPostInSearchResults(t, p1.Id, results.Posts)
			th.checkPostInSearchResults(t, p2.Id, results.Posts)
		})
	})
	t.Run("Should be able to search cyrillic words", func(t *testing.T) {
		p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "слово test", "", 0, false)
		require.Nil(t, err)
		defer th.deleteUserPosts(th.User.Id)

		t.Run("Should search one word", func(t *testing.T) {
			params := &model.SearchParams{Terms: "слово"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 1)
			th.checkPostInSearchResults(t, p1.Id, results.Posts)
		})
		t.Run("Should search using wildcard", func(t *testing.T) {
			params := &model.SearchParams{Terms: "слов*"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 1)
			th.checkPostInSearchResults(t, p1.Id, results.Posts)
		})
	})

	t.Run("Should be able to search japanese words", func(t *testing.T) {
		p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "本", "", 0, false)
		require.Nil(t, err)
		p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "本木", "", 0, false)
		require.Nil(t, err)
		defer th.deleteUserPosts(th.User.Id)

		t.Run("Should search one word", func(t *testing.T) {
			params := &model.SearchParams{Terms: "本"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 2)
			th.checkPostInSearchResults(t, p1.Id, results.Posts)
			th.checkPostInSearchResults(t, p2.Id, results.Posts)
		})
		t.Run("Should search two words", func(t *testing.T) {
			params := &model.SearchParams{Terms: "本木"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 1)
			th.checkPostInSearchResults(t, p2.Id, results.Posts)
		})
		t.Run("Should search with wildcard", func(t *testing.T) {
			params := &model.SearchParams{Terms: "本*"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 2)
			th.checkPostInSearchResults(t, p1.Id, results.Posts)
			th.checkPostInSearchResults(t, p2.Id, results.Posts)
		})
	})

	t.Run("Should be able to search korean words", func(t *testing.T) {
		p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "불", "", 0, false)
		require.Nil(t, err)
		p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "불다", "", 0, false)
		require.Nil(t, err)
		defer th.deleteUserPosts(th.User.Id)

		t.Run("Should search one word", func(t *testing.T) {
			params := &model.SearchParams{Terms: "불"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 1)
			th.checkPostInSearchResults(t, p1.Id, results.Posts)
		})
		t.Run("Should search two words", func(t *testing.T) {
			params := &model.SearchParams{Terms: "불다"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 1)
			th.checkPostInSearchResults(t, p2.Id, results.Posts)
		})
		t.Run("Should search with wildcard", func(t *testing.T) {
			params := &model.SearchParams{Terms: "불*"}
			results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
			require.Nil(t, apperr)

			require.Len(t, results.Posts, 2)
			th.checkPostInSearchResults(t, p1.Id, results.Posts)
			th.checkPostInSearchResults(t, p2.Id, results.Posts)
		})
	})
}

func testSearchAlternativeSpellings(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "Straße test", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "Strasse test", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{Terms: "Straße"}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 2)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
	th.checkPostInSearchResults(t, p2.Id, results.Posts)

	params = &model.SearchParams{Terms: "Strasse"}
	results, apperr = th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 2)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
	th.checkPostInSearchResults(t, p2.Id, results.Posts)
}

func testSearchAlternativeSpellingsAccents(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "café", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "café", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{Terms: "café"}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 2)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
	th.checkPostInSearchResults(t, p2.Id, results.Posts)

	params = &model.SearchParams{Terms: "café"}
	results, apperr = th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 2)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
	th.checkPostInSearchResults(t, p2.Id, results.Posts)

	params = &model.SearchParams{Terms: "cafe"}
	results, apperr = th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 0)
}

func testSearchOrExcludePostsBySpecificUser(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelPrivate.Id, "test fromuser", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User2.Id, th.ChannelPrivate.Id, "test fromuser 2", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)
	defer th.deleteUserPosts(th.User2.Id)

	params := &model.SearchParams{
		Terms:     "fromuser",
		FromUsers: []string{th.User.Id},
	}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 1)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
}

func testSearchOrExcludePostsInChannel(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "test fromuser", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User2.Id, th.ChannelPrivate.Id, "test fromuser 2", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)
	defer th.deleteUserPosts(th.User2.Id)

	params := &model.SearchParams{
		Terms:      "fromuser",
		InChannels: []string{th.ChannelBasic.Id},
	}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 1)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
}

func testSearchOrExcludePostsInDMGM(t *testing.T, th *SearchTestHelper) {
	direct, err := th.createDirectChannel(th.Team.Id, "direct", "direct", []*model.User{th.User, th.User2})
	require.Nil(t, err)
	defer th.deleteChannel(direct)

	group, err := th.createGroupChannel(th.Team.Id, "test group", []*model.User{th.User, th.User2})
	require.Nil(t, err)
	defer th.deleteChannel(group)

	p1, err := th.createPost(th.User.Id, direct.Id, "test fromuser", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User2.Id, group.Id, "test fromuser 2", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)
	defer th.deleteUserPosts(th.User2.Id)

	t.Run("Should be able to search in both DM and GM channels", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:      "fromuser",
			InChannels: []string{direct.Id, group.Id},
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 2)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})

	t.Run("Should be able to search only in DM channel", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:      "fromuser",
			InChannels: []string{direct.Id},
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should be able to search only in GM channel", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:      "fromuser",
			InChannels: []string{group.Id},
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})
}

func testFilterMessagesInSpecificDate(t *testing.T, th *SearchTestHelper) {
	creationDate := model.GetMillisForTime(time.Date(2020, 03, 22, 12, 0, 0, 0, time.UTC))
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "test in specific date", "", creationDate, false)
	require.Nil(t, err)
	creationDate2 := model.GetMillisForTime(time.Date(2020, 03, 23, 0, 0, 0, 0, time.UTC))
	p2, err := th.createPost(th.User.Id, th.ChannelPrivate.Id, "test in the present", "", creationDate2, false)
	require.Nil(t, err)
	creationDate3 := model.GetMillisForTime(time.Date(2020, 03, 21, 23, 59, 59, 0, time.UTC))
	p3, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "test in the present", "", creationDate3, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should be able to search posts on date", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:  "test",
			OnDate: "2020-03-22",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})
	t.Run("Should be able to exclude posts on date", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:        "test",
			ExcludedDate: "2020-03-22",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 2)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
		th.checkPostInSearchResults(t, p3.Id, results.Posts)
	})
}

func testFilterMessagesBeforeSpecificDate(t *testing.T, th *SearchTestHelper) {
	creationDate := model.GetMillisForTime(time.Date(2020, 03, 01, 12, 0, 0, 0, time.UTC))
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "test in specific date", "", creationDate, false)
	require.Nil(t, err)
	creationDate2 := model.GetMillisForTime(time.Date(2020, 03, 22, 23, 59, 59, 0, time.UTC))
	p2, err := th.createPost(th.User.Id, th.ChannelPrivate.Id, "test in specific date 2", "", creationDate2, false)
	require.Nil(t, err)
	creationDate3 := model.GetMillisForTime(time.Date(2020, 03, 26, 16, 55, 0, 0, time.UTC))
	p3, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "test in the present", "", creationDate3, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should be able to search posts before a date", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:      "test",
			BeforeDate: "2020-03-23",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 2)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})

	t.Run("Should be able to exclude posts before a date", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:              "test",
			ExcludedBeforeDate: "2020-03-23",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p3.Id, results.Posts)
	})
}

func testFilterMessagesAfterSpecificDate(t *testing.T, th *SearchTestHelper) {
	creationDate := model.GetMillisForTime(time.Date(2020, 03, 01, 12, 0, 0, 0, time.UTC))
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "test in specific date", "", creationDate, false)
	require.Nil(t, err)
	creationDate2 := model.GetMillisForTime(time.Date(2020, 03, 22, 23, 59, 59, 0, time.UTC))
	p2, err := th.createPost(th.User.Id, th.ChannelPrivate.Id, "test in specific date 2", "", creationDate2, false)
	require.Nil(t, err)
	creationDate3 := model.GetMillisForTime(time.Date(2020, 03, 26, 16, 55, 0, 0, time.UTC))
	p3, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "test in the present", "", creationDate3, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should be able to search posts after a date", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:     "test",
			AfterDate: "2020-03-23",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p3.Id, results.Posts)
	})

	t.Run("Should be able to exclude posts after a date", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:             "test",
			ExcludedAfterDate: "2020-03-23",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 2)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})
}

func testFilterMessagesWithATerm(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "one two three", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelPrivate.Id, "one four five six", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "one seven eight nine", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should exclude terms", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:         "one",
			ExcludedTerms: "five eight",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		th.checkMatchesEqual(t, map[string][]string{
			p1.Id: {"one"},
		}, results.Matches)
	})

	t.Run("Should exclude quoted terms", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:         "one",
			ExcludedTerms: "\"eight nine\"",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		th.checkMatchesEqual(t, map[string][]string{
			p1.Id: {"one"},
			p2.Id: {"one"},
		}, results.Matches)
	})
}

func testSearchUsingBooleanOperators(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "one two three message", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "two messages", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "another message", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should search posts using OR operator", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:   "one two",
			OrTerms: true,
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 2)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})

	t.Run("Should search posts using AND operator", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:   "one two",
			OrTerms: false,
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})
}

func testSearchUsingCombinedFilters(t *testing.T, th *SearchTestHelper) {
	creationDate := model.GetMillisForTime(time.Date(2020, 03, 01, 12, 0, 0, 0, time.UTC))
	p1, err := th.createPost(th.User.Id, th.ChannelPrivate.Id, "one two three message", "", creationDate, false)
	require.Nil(t, err)
	creationDate2 := model.GetMillisForTime(time.Date(2020, 03, 10, 12, 0, 0, 0, time.UTC))
	p2, err := th.createPost(th.User2.Id, th.ChannelPrivate.Id, "two messages", "", creationDate2, false)
	require.Nil(t, err)
	creationDate3 := model.GetMillisForTime(time.Date(2020, 03, 20, 12, 0, 0, 0, time.UTC))
	p3, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "two another message", "", creationDate3, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)
	defer th.deleteUserPosts(th.User2.Id)

	t.Run("Should search combining from user and in channel filters", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:      "two",
			FromUsers:  []string{th.User2.Id},
			InChannels: []string{th.ChannelPrivate.Id},
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})

	t.Run("Should search combining excluding users and in channel filters", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:         "two",
			ExcludedUsers: []string{th.User2.Id},
			InChannels:    []string{th.ChannelPrivate.Id},
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should search combining excluding dates and in channel filters", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:              "two",
			ExcludedBeforeDate: "2020-03-09",
			ExcludedAfterDate:  "2020-03-11",
			InChannels:         []string{th.ChannelPrivate.Id},
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})
	t.Run("Should search combining excluding dates and in channel filters", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:            "two",
			AfterDate:        "2020-03-11",
			ExcludedChannels: []string{th.ChannelPrivate.Id},
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p3.Id, results.Posts)
	})
}

func testSearchIgnoringStopWords(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "the search for a bunch of stop words", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "the objective is to avoid a bunch of stop words", "", 0, false)
	require.Nil(t, err)
	p3, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "in the a on to where you", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should avoid stop word 'the'", func(t *testing.T) {
		params := &model.SearchParams{
			Terms: "the search",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should avoid stop word 'a'", func(t *testing.T) {
		params := &model.SearchParams{
			Terms: "a avoid",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})

	t.Run("Should avoid stop word 'in'", func(t *testing.T) {
		params := &model.SearchParams{
			Terms: "in where",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p3.Id, results.Posts)
	})
}

func testSupportStemming(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "search post", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching post", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "another post", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{
		Terms: "search",
	}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 2)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
	th.checkPostInSearchResults(t, p2.Id, results.Posts)
}

func testSupportWildcards(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "search post", "", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching post", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "another post", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{
		Terms: "search*",
	}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 2)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
	th.checkPostInSearchResults(t, p2.Id, results.Posts)
}

func testNotSupportPrecedingWildcards(t *testing.T, th *SearchTestHelper) {
	_, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "search post", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "searching post", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "another post", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{
		Terms: "*earch",
	}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 0)
}

func testSearchDiscardWildcardAlone(t *testing.T, th *SearchTestHelper) {
	_, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "search post", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "searching post", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "another post", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{
		Terms: "search *",
	}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 0)
}

func testSupportTermsWithDash(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "search term-with-dash", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "searching term with dash", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should search terms with dash", func(t *testing.T) {
		params := &model.SearchParams{
			Terms: "term-with-dash",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should search terms with dash using quotes", func(t *testing.T) {
		params := &model.SearchParams{
			Terms: "\"term-with-dash\"",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})
}

func testSupportTermsWithUnderscore(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "search term_with_underscore", "", 0, false)
	require.Nil(t, err)
	_, err = th.createPost(th.User.Id, th.ChannelBasic.Id, "searching term with underscore", "", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should search terms with underscore", func(t *testing.T) {
		params := &model.SearchParams{
			Terms: "term_with_underscore",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should search terms with underscore using quotes", func(t *testing.T) {
		params := &model.SearchParams{
			Terms: "\"term_with_underscore\"",
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})
}

func testSearchOrExcludePostsWithHashtags(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "search post with #hashtag", "#hashtag", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching term with hashtag", "", 0, false)
	require.Nil(t, err)
	p3, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching term with", "#hashtag", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should search terms with hashtags", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:     "#hashtag",
			IsHashtag: true,
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 2)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p3.Id, results.Posts)
	})

	t.Run("Should search hashtag terms without hashtag option", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:     "#hashtag",
			IsHashtag: false,
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 2)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})
}

func testSearchHashtagWithMarkdown(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching hashtag #hashtag", "#hashtag", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching term with `#hashtag`", "#hashtag", 0, false)
	require.Nil(t, err)
	p3, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching term with **#hashtag**", "#hashtag", 0, false)
	require.Nil(t, err)
	p4, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching term with ~~#hashtag~~", "#hashtag", 0, false)
	require.Nil(t, err)
	p5, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching term with _#hashtag_", "#hashtag", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{
		Terms:     "#hashtag",
		IsHashtag: true,
	}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 5)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
	th.checkPostInSearchResults(t, p2.Id, results.Posts)
	th.checkPostInSearchResults(t, p3.Id, results.Posts)
	th.checkPostInSearchResults(t, p4.Id, results.Posts)
	th.checkPostInSearchResults(t, p5.Id, results.Posts)
}

func testSearcWithMultipleHashtags(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching hashtag #hashtag", "#hashtwo #hashone", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching term with `#hashtag`", "#hashtwo #hashthree", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Should search posts with multiple hashtags", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:     "#hashone #hashtwo",
			IsHashtag: true,
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 1)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
	})

	t.Run("Should search posts with multiple hashtags using OR", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:     "#hashone #hashtwo",
			IsHashtag: true,
			OrTerms:   true,
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 2)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
	})
}

func testSearchPostsWithDotsInHashtags(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching hashtag #hashtag.dot", "#hashtag.dot", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	params := &model.SearchParams{
		Terms:     "#hashtag.dot",
		IsHashtag: true,
	}
	results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
	require.Nil(t, apperr)

	require.Len(t, results.Posts, 1)
	th.checkPostInSearchResults(t, p1.Id, results.Posts)
}

func testSearchHashtagCaseInsensitive(t *testing.T, th *SearchTestHelper) {
	p1, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching hashtag #HaShTaG", "#HaShTaG", 0, false)
	require.Nil(t, err)
	p2, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching hashtag #hashtag", "#hashtag", 0, false)
	require.Nil(t, err)
	p3, err := th.createPost(th.User.Id, th.ChannelBasic.Id, "searching hashtag #HASHTAG", "#HASHTAG", 0, false)
	require.Nil(t, err)
	defer th.deleteUserPosts(th.User.Id)

	t.Run("Lower case hashtag", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:     "#hashtag",
			IsHashtag: true,
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 3)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
		th.checkPostInSearchResults(t, p3.Id, results.Posts)
	})

	t.Run("Upper case hashtag", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:     "#HASHTAG",
			IsHashtag: true,
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 3)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
		th.checkPostInSearchResults(t, p3.Id, results.Posts)
	})

	t.Run("Mixed case hashtag", func(t *testing.T) {
		params := &model.SearchParams{
			Terms:     "#HaShTaG",
			IsHashtag: true,
		}
		results, apperr := th.Store.Post().SearchPostsInTeamForUser([]*model.SearchParams{params}, th.User.Id, th.Team.Id, false, false, 0, 20)
		require.Nil(t, apperr)

		require.Len(t, results.Posts, 3)
		th.checkPostInSearchResults(t, p1.Id, results.Posts)
		th.checkPostInSearchResults(t, p2.Id, results.Posts)
		th.checkPostInSearchResults(t, p3.Id, results.Posts)
	})
}
