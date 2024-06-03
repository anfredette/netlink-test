use inotify::{Inotify, WatchMask};
use std::ffi::OsStr;
use std::path::Path;

fn main() {
    println!("Hello, world!");
    println!("Calling nsWatcher");
    //ns_watcher();
    int_watcher();
}

fn _ns_watcher() {
    let mut inotify = Inotify::init().expect("Failed to initialize inotify");
    let watch_path = Path::new("/var/run/netns");

    inotify
        .watches()
        .add(watch_path, WatchMask::CREATE | WatchMask::DELETE)
        .expect("Failed to add inotify watch");

    let mut buffer = [0; 1024];
    loop {
        let events = inotify
            .read_events_blocking(&mut buffer)
            .expect("Failed to read inotify events");

        for event in events {
            println!(
                "Network namespace name: {:?}, event: {:?}",
                event.name.unwrap_or(OsStr::new("None")),
                event.mask
            );
        }
    }
}

use netlink_packet_core::{NetlinkMessage, NetlinkPayload};
use netlink_packet_route::link::LinkMessage;
use netlink_sys::{protocols::NETLINK_ROUTE, Socket, SocketAddr};

fn int_watcher() {
    let mut socket = Socket::new(NETLINK_ROUTE).unwrap();
    let sa = SocketAddr::new(0, 1);
    socket.bind(&sa).unwrap();

    let mut buf = vec![0; 4096];
    loop {
        let size = socket.recv(&mut buf, 0).unwrap();
        //let packet = NetlinkMessage::deserialize(&buf[..size]).unwrap();
        let packet = NetlinkMessage::<LinkMessage>::deserialize(&buf[..size]).unwrap();

        let message_type = packet.header.message_type;

        let (header, payload) = packet.into_parts();

        println!("Message type: {}, Header: {:?}", message_type, header);

        //     // if let NetlinkMessage::NewLink(link) = packet.payload {
        //     //     handle_link_event(&link);
        //     // }
        //     if let LinkMessage {
        //         header: link,
        //         attributes,
        //     } = packet.payload
        //     {
        //         handle_link_event(&link);
        //     }
        // }
    }
}

// fn handle_link_event(link: &LinkMessage) {
//     for nla in &link.nlas {
//         match nla {
//             Nla::IfName(name) => println!("Interface name: {}", name),
//             Nla::OperState(state) => match state {
//                 0 => println!("Interface is down"),
//                 6 => println!("Interface is up"),
//                 _ => (),
//             },
//             _ => (),
//         }
//     }
// }
