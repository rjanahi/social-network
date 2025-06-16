import Head from 'next/head';
import { useEffect } from 'react';

export default function Chat() {
  useEffect(() => {
    // Load external scripts after mount
    const script1 = document.createElement("script");
    script1.src = "/js/chat.js";
    script1.defer = true;
    document.body.appendChild(script1);

    const script2 = document.createElement("script");
    script2.src = "/js/socket.js";
    script2.defer = true;
    document.body.appendChild(script2);

    const script3 = document.createElement("script");
    script3.src = "/js/session.js";
    script3.defer = true;
    document.body.appendChild(script3);
  }, []);

  return (
    <>
      <Head>
        <title>Chat</title>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <link rel="stylesheet" href="/css/style.css" />
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
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
              <button className="return-button" onClick={() => window.location.href = '/posts'}>Return</button>
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
