import Head from 'next/head';

export default function Chat() {
  return (
    <>
      <Head>
        <title>Chat</title>
        <link rel="stylesheet" href="/css/style.css" />
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
        <script src="/js/chat.js" defer></script>
        <script src="/js/socket.js" defer></script>
        <script src="/js/session.js" defer></script>
      </Head>
      <section id="chatSection">
        <div className="chat-container">
          <div className="chat-sidebar">
            <button className="create-group" id="createGroupButton">Create Group</button>
            <h3>Users</h3>
            <ul id="userList"></ul>
          </div>
          <div id="group-container" hidden>
            <h3>Create Group</h3>
            <form id="createGroupForm">
              <input type="text" id="groupName" placeholder="Group Name" required />
              <div id="groupMembers" className="user-checkbox-list"></div>
              <button type="submit" className="button-main">Create Group</button>
            </form>
          </div>
          <div className="chat-main" id="chat-main">
            <div className="chat-header">
              <button className="return-button" onClick={() => window.location.href='/posts.html'}>Return</button>
              <h3 id="chatWithLabel">Chat</h3>
            </div>
            <div id="chatWindow" className="chat-window">
              <div id="messages" className="messages-container"></div>
            </div>
            <form id="chatForm">
              <input type="text" id="chatInput" placeholder="Type a message..." required />
              <button type="submit" className="button-main">Send</button>
            </form>
          </div>
        </div>
      </section>
    </>
  );
}