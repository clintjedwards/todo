use crate::config;
use ::tui::layout::{Constraint, Direction, Layout};
use ::tui::widgets::{Block, Borders, Widget};
use ::tui::{backend::TermionBackend, Terminal};
use anyhow::Result;
use std::sync::mpsc;
use std::thread;
use std::time::{Duration, Instant};
use std::{
    error::Error,
    io,
    sync::mpsc::{Receiver, Sender},
};
use termion::input::TermRead;
use termion::{clear, event::Key, input::MouseTerminal, raw::IntoRawMode, screen::AlternateScreen};

pub struct TUI {
    host: String,
    redraw_rate: Duration,
}

pub fn new() -> TUI {
    let config = config::get_cli();
    let config = match config {
        Ok(config) => config,
        Err(error) => panic!("Error reading environment variable: {}", error),
    };

    TUI {
        host: config.host,
        redraw_rate: Duration::from_millis(200),
    }
}

enum Event {
    Input(Key),
    Redraw,
}

impl TUI {
    // We first need to draw the Tui and then block for an event, this event can be a redraw
    // or a user action
    pub fn start(&self) -> Result<()> {
        let stdout = io::stdout().into_raw_mode()?;
        let stdout = AlternateScreen::from(stdout);
        let stdout = AlternateScreen::from(stdout);
        let backend = TermionBackend::new(stdout);
        let mut terminal = Terminal::new(backend)?;
        terminal.clear()?;

        let (send_chan, receive_chan) = mpsc::channel();

        start_event_emitter(send_chan.clone());
        start_redraw_emitter(send_chan.clone(), self.redraw_rate);

        // Main event loop. We first draw the frame and then listen to detect what to do next.
        // If it's a simple redraw command we redraw and move on, if its a key event we handle
        // the key.
        loop {
            terminal.draw(|frame| {
                let layout = Layout::default()
                    .direction(Direction::Vertical)
                    .constraints([Constraint::Percentage(100)].as_ref())
                    .split(frame.size());

                let block = Block::default().title("Todo list");
                frame.render_widget(block, layout[0]);
            })?;

            match receive_chan.recv()? {
                Event::Input(key) => {
                    match key {
                        Key::Char('q') => break,
                        Key::Ctrl('c') => break,
                        _ => {}
                    };
                }
                Event::Redraw => {}
            }
        }

        Ok(())
    }
}

// start_event_emitter starts a new thread which polls for user key presses and sends them
// to the provided event handler
fn start_event_emitter(sender: Sender<Event>) {
    thread::spawn(move || {
        let stdin = io::stdin();

        for event in stdin.keys() {
            if let Ok(key) = event {
                if let Err(err) = sender.send(Event::Input(key)) {
                    eprintln!("{}", err);
                    return;
                }
            }
        }
    });
}

// start_redraw_emitter starts a new thread which sends a redraw events and then sleeps for the appropriate amount of time.
fn start_redraw_emitter(sender: Sender<Event>, duration: Duration) {
    thread::spawn(move || loop {
        if sender.send(Event::Redraw).is_err() {
            break;
        }
        thread::sleep(duration);
    });
}
