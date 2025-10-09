"use client";

import { useState } from "react";

export function TelegramSetupGuide() {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="rounded-2xl border border-border/40 bg-card/40 backdrop-blur-xl overflow-hidden">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-6 py-4 flex items-center justify-between text-left hover:bg-card/60 transition-colors"
      >
        <div className="flex items-center gap-3">
          <span className="text-2xl">💡</span>
          <div>
            <h3 className="text-lg font-semibold text-foreground">
              How to Get Your Telegram Chat ID
            </h3>
            <p className="text-sm text-muted-foreground">
              Quick 3-step guide to start receiving notifications
            </p>
          </div>
        </div>
        <svg
          className={`w-5 h-5 text-muted-foreground transition-transform ${
            isExpanded ? "rotate-180" : ""
          }`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>

      {isExpanded && (
        <div className="px-6 pb-6 space-y-6">
          <div className="h-px bg-border/40" />

          {/* Step 1 */}
          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 rounded-full bg-[#0088cc]/20 border border-[#0088cc]/30 flex items-center justify-center text-[#0088cc] font-semibold">
              1
            </div>
            <div className="flex-1 space-y-2">
              <h4 className="font-semibold text-foreground">
                Open Telegram and search for @userinfobot
              </h4>
              <p className="text-sm text-muted-foreground">
                This is an official bot that will provide your chat ID. Simply
                search for it in Telegram's search bar.
              </p>
            </div>
          </div>

          {/* Step 2 */}
          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 rounded-full bg-[#0088cc]/20 border border-[#0088cc]/30 flex items-center justify-center text-[#0088cc] font-semibold">
              2
            </div>
            <div className="flex-1 space-y-2">
              <h4 className="font-semibold text-foreground">
                Start the bot and send any message
              </h4>
              <p className="text-sm text-muted-foreground">
                Click "Start" or send any message to @userinfobot. It will
                immediately reply with your user information.
              </p>
            </div>
          </div>

          {/* Step 3 */}
          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 rounded-full bg-[#0088cc]/20 border border-[#0088cc]/30 flex items-center justify-center text-[#0088cc] font-semibold">
              3
            </div>
            <div className="flex-1 space-y-2">
              <h4 className="font-semibold text-foreground">
                Copy your Chat ID and paste it below
              </h4>
              <p className="text-sm text-muted-foreground">
                The bot will show your ID as a number (e.g., 123456789). Copy
                this number and add it using the form below.
              </p>
              <div className="mt-3 p-3 rounded-lg bg-muted/40 border border-border/30">
                <p className="text-xs text-muted-foreground mb-1">
                  Example response:
                </p>
                <code className="text-sm text-foreground font-mono">
                  Id: 123456789
                  <br />
                  First name: John
                  <br />
                  Username: @johndoe
                </code>
              </div>
            </div>
          </div>

          {/* Additional Info */}
          <div className="mt-6 p-4 rounded-lg bg-blue-500/10 border border-blue-500/20">
            <div className="flex gap-3">
              <span className="text-blue-500 flex-shrink-0">ℹ️</span>
              <div className="text-sm text-blue-200/90">
                <p className="font-semibold mb-1">Important Note</p>
                <p>
                  Make sure the LunaSentri bot has been configured with a valid
                  Telegram Bot Token by your administrator. You'll receive a test
                  message when you click the "Test" button after adding your chat
                  ID.
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
